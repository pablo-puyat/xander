// Package processor orchestrates the comic parsing and matching workflow.
// It coordinates between LLM parsing, ComicVine searches, and batch processing.
package processor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"comic-parser/comicvine"
	"comic-parser/config"
	"comic-parser/llm"
	"comic-parser/models"
	"comic-parser/parser"
	"comic-parser/selector"
)

// LLMClient defines the interface for LLM interactions.
type LLMClient interface {
	CompleteWithRetry(ctx context.Context, prompt string, maxRetries int, delay time.Duration) (string, error)
	Close()
}

// CVClient defines the interface for ComicVine interactions.
type CVClient interface {
	SearchIssues(ctx context.Context, title string, issueNumber string) ([]models.ComicVineIssue, error)
	Close()
}

// Processor orchestrates the comic parsing and matching workflow.
type Processor struct {
	cfg       *config.Config
	llmClient LLMClient
	parser    parser.Parser
	cvClient  CVClient
	selector  selector.Selector
	verbose   bool

	// Progress tracking
	progressMu sync.Mutex
	progress   models.BatchProgress
}

// NewProcessor creates a new processor.
func NewProcessor(cfg *config.Config) *Processor {
	llmClient := llm.NewClient(cfg)

	var sel selector.Selector
	if cfg.Interactive {
		sel = selector.NewTUISelector()
	} else {
		sel = selector.NewLLMSelector(llmClient, cfg)
	}

	// Setup parser pipeline
	regexParser := parser.NewRegexParser()
	llmParser := parser.NewLLMParser(llmClient, cfg)
	pipeline := parser.NewPipelineParser(regexParser, llmParser)

	return &Processor{
		cfg:       cfg,
		llmClient: llmClient,
		parser:    pipeline,
		cvClient:  comicvine.NewClient(cfg),
		selector:  sel,
		verbose:   cfg.Verbose,
	}
}

// Close cleans up processor resources.
func (p *Processor) Close() {
	if p.cvClient != nil {
		p.cvClient.Close()
	}
	if p.llmClient != nil {
		p.llmClient.Close()
	}
}

// ProcessFile processes a single comic filename.
// It returns a ProcessingResult containing match information or an error description.
func (p *Processor) ProcessFile(ctx context.Context, filename string) (*models.ProcessingResult, error) {
	startTime := time.Now()

	result := &models.ProcessingResult{
		Filename:    filename,
		ProcessedAt: startTime,
	}

	// Step 1: Parse the filename using Parser Pipeline
	if p.verbose {
		log.Printf("Parsing filename: %s", filename)
	}

	input := models.ParsedFilename{OriginalFilename: filename}
	parsedVal, err := p.parser.Parse(ctx, input)
	parsed := &parsedVal

	if err != nil {
		result.Error = fmt.Sprintf("parsing filename: %v", err)
		result.ProcessingTimeMS = time.Since(startTime).Milliseconds()
		return result, nil
	}

	if p.verbose {
		log.Printf("Parsed: title=%q issue=%q year=%q", parsed.Title, parsed.IssueNumber, parsed.Year)
	}

	// Step 2: Search ComicVine
	if p.verbose {
		log.Printf("Searching ComicVine for: %s #%s", parsed.Title, parsed.IssueNumber)
	}

	issues, err := p.cvClient.SearchIssues(ctx, parsed.Title, parsed.IssueNumber)
	if err != nil {
		result.Error = fmt.Sprintf("searching comicvine: %v", err)
		result.ProcessingTimeMS = time.Since(startTime).Milliseconds()
		return result, nil
	}

	if p.verbose {
		log.Printf("Found %d results from ComicVine", len(issues))
	}

	// Step 3: Match results using Selector
	match, err := p.selector.Select(ctx, parsed, issues)
	if err != nil {
		result.Error = fmt.Sprintf("matching results: %v", err)
		result.ProcessingTimeMS = time.Since(startTime).Milliseconds()
		return result, nil
	}

	result.Success = true
	result.Match = match
	result.ProcessingTimeMS = time.Since(startTime).Milliseconds()

	if p.verbose {
		if match.SelectedIssue != nil {
			log.Printf("Matched: %s #%s (%s) - %s",
				match.SelectedIssue.Volume.Name,
				match.SelectedIssue.IssueNumber,
				match.MatchConfidence,
				match.ComicVineURL)
		} else {
			log.Printf("No match found: %s", match.Reasoning)
		}
	}

	return result, nil
}

// ProcessBatch processes multiple files concurrently using a worker pool.
// Results are sent to the provided channel as they complete.
func (p *Processor) ProcessBatch(ctx context.Context, filenames []string, resultChan chan<- *models.ProcessingResult) {
	p.progress = models.BatchProgress{
		Total: len(filenames),
	}

	// Create worker pool
	jobs := make(chan string, len(filenames))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < p.cfg.WorkerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for filename := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				result, _ := p.ProcessFile(ctx, filename)

				p.progressMu.Lock()
				p.progress.Processed++
				if result.Success {
					p.progress.Successful++
				} else {
					p.progress.Failed++
				}
				p.progressMu.Unlock()

				resultChan <- result
			}
		}(i)
	}

	// Send jobs
	for _, filename := range filenames {
		jobs <- filename
	}
	close(jobs)

	// Wait for completion
	wg.Wait()
}

// GetProgress returns the current processing progress in a thread-safe manner.
func (p *Processor) GetProgress() models.BatchProgress {
	p.progressMu.Lock()
	defer p.progressMu.Unlock()
	return p.progress
}

