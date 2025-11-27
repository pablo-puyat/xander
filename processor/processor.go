// Package processor orchestrates the comic parsing and matching workflow.
// It coordinates between LLM parsing, ComicVine searches, and batch processing.
package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"comic-parser/comicvine"
	"comic-parser/config"
	"comic-parser/llm"
	"comic-parser/models"
	"comic-parser/prompts"
)

// Processor orchestrates the comic parsing and matching workflow.
type Processor struct {
	cfg         *config.Config
	llmClient   *llm.Client
	cvClient    *comicvine.Client
	verbose     bool
	
	// Progress tracking
	progressMu sync.Mutex
	progress   models.BatchProgress
}

// NewProcessor creates a new processor.
func NewProcessor(cfg *config.Config) *Processor {
	return &Processor{
		cfg:       cfg,
		llmClient: llm.NewClient(cfg),
		cvClient:  comicvine.NewClient(cfg),
		verbose:   cfg.Verbose,
	}
}

// Close cleans up processor resources.
func (p *Processor) Close() {
	p.cvClient.Close()
}

// ProcessFile processes a single comic filename.
// It returns a ProcessingResult containing match information or an error description.
func (p *Processor) ProcessFile(ctx context.Context, filename string) (*models.ProcessingResult, error) {
	startTime := time.Now()

	result := &models.ProcessingResult{
		Filename:    filename,
		ProcessedAt: startTime,
	}

	// Step 1: Parse the filename using LLM
	if p.verbose {
		log.Printf("Parsing filename: %s", filename)
	}

	parsed, err := p.parseFilename(ctx, filename)
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

	// Step 3: Match results using LLM
	match, err := p.matchResults(ctx, parsed, issues)
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

// parseFilename uses the LLM to parse a comic filename
func (p *Processor) parseFilename(ctx context.Context, filename string) (*models.ParsedFilename, error) {
	prompt := prompts.FilenameParsePrompt(filename)

	response, err := p.llmClient.CompleteWithRetry(
		ctx,
		prompt,
		p.cfg.RetryAttempts,
		time.Duration(p.cfg.RetryDelaySeconds)*time.Second,
	)
	if err != nil {
		return nil, fmt.Errorf("LLM completion: %w", err)
	}

	// Extract JSON from response
	jsonStr := llm.ExtractJSON(response)

	var parsed models.ParsedFilename
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w (response: %s)", err, response)
	}

	parsed.OriginalFilename = filename
	return &parsed, nil
}

// matchResults uses the LLM to select the best match from ComicVine results
func (p *Processor) matchResults(ctx context.Context, parsed *models.ParsedFilename, issues []models.ComicVineIssue) (*models.MatchResult, error) {
	result := &models.MatchResult{
		OriginalFilename: parsed.OriginalFilename,
		ParsedInfo:       *parsed,
	}

	if len(issues) == 0 {
		result.MatchConfidence = "none"
		result.Reasoning = "No results found in ComicVine"
		return result, nil
	}

	prompt := prompts.ResultMatchPrompt(*parsed, issues)

	response, err := p.llmClient.CompleteWithRetry(
		ctx,
		prompt,
		p.cfg.RetryAttempts,
		time.Duration(p.cfg.RetryDelaySeconds)*time.Second,
	)
	if err != nil {
		return nil, fmt.Errorf("LLM completion: %w", err)
	}

	// Extract JSON from response
	jsonStr := llm.ExtractJSON(response)

	var matchResp prompts.MatchResponse
	if err := json.Unmarshal([]byte(jsonStr), &matchResp); err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w (response: %s)", err, response)
	}

	result.MatchConfidence = matchResp.MatchConfidence
	result.Reasoning = matchResp.Reasoning

	if matchResp.SelectedIndex >= 0 && matchResp.SelectedIndex < len(issues) {
		selectedIssue := issues[matchResp.SelectedIndex]
		result.SelectedIssue = &selectedIssue
		result.ComicVineID = selectedIssue.ID
		result.ComicVineURL = selectedIssue.SiteDetailURL
	}

	return result, nil
}
