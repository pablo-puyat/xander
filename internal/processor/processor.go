// Package processor orchestrates the comic parsing and matching workflow.
// It coordinates between LLM parsing, ComicVine searches, and batch processing.
package processor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"comic-parser/internal/config"
	"comic-parser/internal/models"
	"comic-parser/internal/parser"
	"comic-parser/internal/selector"
	"comic-parser/internal/storage"
)

// CVClient defines the interface for ComicVine interactions.
type CVClient interface {
	SearchIssues(ctx context.Context, title string, issueNumber string) ([]models.ComicVineIssue, error)
	Close()
}

// Processor orchestrates the comic parsing and matching workflow.
type Processor struct {
	cfg      *config.Config
	parser   parser.Parser
	cvClient CVClient
	selector selector.Selector
	store    *storage.Storage
	verbose  bool

	// Progress tracking
	progressMu sync.Mutex
	progress   models.BatchProgress
}

// NewProcessor creates a new processor.
func NewProcessor(cfg *config.Config, p parser.Parser, cvClient CVClient, sel selector.Selector, store *storage.Storage) *Processor {
	return &Processor{
		cfg:      cfg,
		parser:   p,
		cvClient: cvClient,
		selector: sel,
		store:    store,
		verbose:  cfg.Verbose,
	}
}

// Close cleans up processor resources.
func (p *Processor) Close() {
	if p.cvClient != nil {
		p.cvClient.Close()
	}
	// Parser is managed externally
}

// ProcessFile processes a single comic filename.
// It returns a ProcessingResult containing match information or an error description.
func (p *Processor) ProcessFile(ctx context.Context, filename string) (*models.ProcessingResult, error) {
	startTime := time.Now()

	result := &models.ProcessingResult{
		Filename:    filename,
		ProcessedAt: startTime,
	}

	// Step 1: Parse the filename
	if p.verbose {
		log.Printf("Parsing filename: %s", filename)
	}

	parsed, err := p.parser.Parse(ctx, &models.ParsedFilename{OriginalFilename: filename})
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

// ParseBatch processes files for parsing only and saves results to the database.
func (p *Processor) ParseBatch(ctx context.Context, filenames []string, parserName string) {
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

				err := p.ProcessFileParseOnly(ctx, filename, parserName)

				p.progressMu.Lock()
				p.progress.Processed++
				if err == nil {
					p.progress.Successful++
				} else {
					p.progress.Failed++
				}
				p.progressMu.Unlock()
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

// ProcessFileParseOnly parses a single file and saves the result to the database.
func (p *Processor) ProcessFileParseOnly(ctx context.Context, filename string, parserName string) error {
	if p.verbose {
		log.Printf("Parsing filename: %s", filename)
	}

	parsed, err := p.parser.Parse(ctx, &models.ParsedFilename{OriginalFilename: filename})
	if err != nil {
		if p.verbose {
			log.Printf("Error parsing %s: %v", filename, err)
		}
		return err
	}

	if p.verbose {
		log.Printf("Parsed: title=%q issue=%q", parsed.Title, parsed.IssueNumber)
	}

	if p.store != nil {
		if err := p.store.SaveParsedFilename(ctx, parsed, parserName); err != nil {
			if p.verbose {
				log.Printf("Error saving parsed result for %s: %v", filename, err)
			}
			return err
		}
	} else {
		log.Printf("Warning: No storage configured, result not saved for %s", filename)
	}

	return nil
}
