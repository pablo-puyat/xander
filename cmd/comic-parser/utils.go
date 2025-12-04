package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"comic-parser/internal/config"
	"comic-parser/internal/models"
	"comic-parser/internal/processor"
	"comic-parser/internal/storage"
)

func processSingle(ctx context.Context, proc *processor.Processor, filename string) {
	fmt.Printf("Processing: %s\n\n", filename)

	result, err := proc.ProcessFile(ctx, filename)
	if err != nil {
		log.Fatalf("Error processing file: %v", err)
	}

	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
		return
	}

	if result.Match == nil {
		fmt.Println("No match result generated")
		return
	}

	fmt.Println("=== Parsed Information ===")
	fmt.Printf("Title:        %s\n", result.Match.ParsedInfo.Title)
	fmt.Printf("Issue:        %s\n", result.Match.ParsedInfo.IssueNumber)
	fmt.Printf("Year:         %s\n", result.Match.ParsedInfo.Year)
	fmt.Printf("Publisher:    %s\n", result.Match.ParsedInfo.Publisher)
	fmt.Printf("Volume:       %s\n", result.Match.ParsedInfo.VolumeNumber)
	fmt.Printf("Confidence:   %s\n", result.Match.ParsedInfo.Confidence)
	fmt.Printf("Notes:        %s\n", result.Match.ParsedInfo.Notes)

	fmt.Println("\n=== ComicVine Match ===")
	if result.Match.SelectedIssue != nil {
		issue := result.Match.SelectedIssue
		fmt.Printf("Series:       %s\n", issue.Volume.Name)
		fmt.Printf("Issue:        #%s\n", issue.IssueNumber)
		fmt.Printf("Cover Date:   %s\n", issue.CoverDate)
		fmt.Printf("Publisher:    %s\n", issue.Volume.Publisher)
		fmt.Printf("ComicVine ID: %d\n", issue.ID)
		fmt.Printf("URL:          %s\n", issue.SiteDetailURL)
	} else {
		fmt.Println("No match found")
	}
	fmt.Printf("Confidence:   %s\n", result.Match.MatchConfidence)
	fmt.Printf("Reasoning:    %s\n", result.Match.Reasoning)
	fmt.Printf("\nProcessing time: %dms\n", result.ProcessingTimeMS)
}

func processBatch(ctx context.Context, proc *processor.Processor, cfg *config.Config, filenames []string) {
	resultChan := make(chan *models.ProcessingResult, 100)
	var results []*models.ProcessingResult

	// Start collecting results
	done := make(chan struct{})
	go func() {
		for result := range resultChan {
			results = append(results, result)

			// Print progress
			progress := proc.GetProgress()
			fmt.Printf("\rProgress: %d/%d (✓ %d, ✗ %d)",
				progress.Processed, progress.Total,
				progress.Successful, progress.Failed)
		}
		close(done)
	}()

	// Start processing
	startTime := time.Now()
	proc.ProcessBatch(ctx, filenames, resultChan)
	close(resultChan)
	<-done

	fmt.Println() // New line after progress

	// Save results
	if err := saveResults(results, cfg.OutputFile, cfg.OutputFormat); err != nil {
		log.Printf("Error saving results: %v", err)
	} else {
		fmt.Printf("\nResults saved to: %s\n", cfg.OutputFile)
	}

	// Print summary
	elapsed := time.Since(startTime)
	progress := proc.GetProgress()
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total processed: %d\n", progress.Processed)
	fmt.Printf("Successful:      %d\n", progress.Successful)
	fmt.Printf("Failed:          %d\n", progress.Failed)
	fmt.Printf("Time elapsed:    %s\n", elapsed.Round(time.Second))
	if progress.Processed > 0 {
		fmt.Printf("Avg time/file:   %s\n", (elapsed / time.Duration(progress.Processed)).Round(time.Millisecond))
	}
}

func loadFilenames(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var filenames []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			filenames = append(filenames, line)
		}
	}

	return filenames, scanner.Err()
}

func saveResults(results []*models.ProcessingResult, path string, format string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	switch format {
	case "json":
		return saveJSON(results, path)
	case "csv":
		return saveCSV(results, path)
	case "sqlite", "db":
		return saveDB(results, path)
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}

func saveDB(results []*models.ProcessingResult, path string) error {
	store, err := storage.NewStorage(path)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	for _, result := range results {
		if err := store.SaveResult(ctx, result); err != nil {
			return fmt.Errorf("failed to save result for %s: %w", result.Filename, err)
		}
	}
	return nil
}

func saveJSON(results []*models.ProcessingResult, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func saveCSV(results []*models.ProcessingResult, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Filename",
		"Success",
		"Error",
		"Parsed_Title",
		"Parsed_Issue",
		"Parsed_Year",
		"Match_Confidence",
		"ComicVine_ID",
		"ComicVine_Series",
		"ComicVine_Issue",
		"ComicVine_CoverDate",
		"ComicVine_Publisher",
		"ComicVine_URL",
		"Reasoning",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write rows
	for _, r := range results {
		row := []string{
			r.Filename,
			fmt.Sprintf("%t", r.Success),
			r.Error,
		}

		if r.Match != nil {
			row = append(row,
				r.Match.ParsedInfo.Title,
				r.Match.ParsedInfo.IssueNumber,
				r.Match.ParsedInfo.Year,
				r.Match.MatchConfidence,
			)

			if r.Match.SelectedIssue != nil {
				row = append(row,
					fmt.Sprintf("%d", r.Match.ComicVineID),
					r.Match.SelectedIssue.Volume.Name,
					r.Match.SelectedIssue.IssueNumber,
					r.Match.SelectedIssue.CoverDate,
					r.Match.SelectedIssue.Volume.Publisher,
					r.Match.ComicVineURL,
				)
			} else {
				row = append(row, "", "", "", "", "", "")
			}
			row = append(row, r.Match.Reasoning)
		} else {
			row = append(row, "", "", "", "", "", "", "", "", "", "", "")
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
