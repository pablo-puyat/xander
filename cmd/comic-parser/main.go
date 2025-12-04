package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"comic-parser/internal/comicvine"
	"comic-parser/internal/config"
	"comic-parser/internal/llm"
	"comic-parser/internal/models"
	"comic-parser/internal/parser"
	"comic-parser/internal/processor"
	"comic-parser/internal/selector"
	"comic-parser/internal/storage"
	"comic-parser/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Define flags
	inputFile := flag.String("input", "", "Input file containing filenames (one per line)")
	outputFile := flag.String("output", "results.json", "Output file for results")
	outputFormat := flag.String("format", "json", "Output format: json, csv, or sqlite")
	configFile := flag.String("config", "config.json", "Path to configuration file")
	workers := flag.Int("workers", 3, "Number of concurrent workers")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	interactive := flag.Bool("interactive", false, "Enable interactive TUI mode")
	singleFile := flag.String("file", "", "Process a single filename (for testing)")
	generateConfig := flag.Bool("generate-config", false, "Generate a sample config file")
	parserName := flag.String("parser", "", "Parser to use: regex or llm (enables parse-only mode)")
	dbPath := flag.String("db", "comics.db", "Database path for storing results")
	tuiMode := flag.Bool("tui", false, "Launch TUI to view parsed results")

	flag.Parse()

	// Handle config generation
	if *generateConfig {
		cfg := config.DefaultConfig()
		cfg.AnthropicAPIKey = "your-anthropic-api-key-here"
		cfg.ComicVineAPIKey = "your-comicvine-api-key-here"
		if err := cfg.SaveConfig("config.sample.json"); err != nil {
			log.Fatalf("Error generating config: %v", err)
		}
		fmt.Println("Generated config.sample.json - copy to config.json and add your API keys")
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	cfg.LoadFromEnv()

	// Override config with flags
	if *workers > 0 {
		cfg.WorkerCount = *workers
	}
	if *outputFile != "" {
		cfg.OutputFile = *outputFile
	}
	if *outputFormat != "" {
		cfg.OutputFormat = *outputFormat
	}
	cfg.Verbose = *verbose
	cfg.Interactive = *interactive

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create shared HTTP client
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Create dependencies
	llmClient := llm.NewClient(cfg, httpClient)
	defer llmClient.Close()

	cvClient := comicvine.NewClient(cfg, httpClient)

	// Create parser
	var p parser.Parser
	if *parserName != "" {
		switch *parserName {
		case "regex":
			p = parser.NewRegexParser()
		case "llm":
			p = parser.NewLLMParser(llmClient, cfg.RetryAttempts, cfg.RetryDelaySeconds)
		default:
			log.Fatalf("Unknown parser: %s (must be regex or llm)", *parserName)
		}
	} else {
		// Since chain parser is removed, we require a parser to be specified
		log.Fatal("Please specify a parser using -parser (regex or llm)")
	}

	// Create selector
	var sel selector.Selector
	if cfg.Interactive {
		sel = selector.NewTUISelector()
	} else {
		sel = selector.NewLLMSelector(llmClient, cfg)
	}

	// Initialize Storage if parsing is enabled or TUI mode
	var store *storage.Storage
	if *parserName != "" || *tuiMode {
		var err error
		store, err = storage.NewStorage(*dbPath)
		if err != nil {
			log.Fatalf("Error initializing storage: %v", err)
		}
		defer store.Close()
	}

	// Create processor
	proc := processor.NewProcessor(cfg, p, cvClient, sel, store)
	defer proc.Close()

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
		cancel()
	}()

	if *tuiMode {
		// Initialize TUI
		model, err := tui.NewModel(ctx, store, cvClient)
		if err != nil {
			log.Fatalf("Error initializing TUI: %v", err)
		}

		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			log.Fatalf("Error running TUI: %v", err)
		}
		return
	}

	// Process single file or batch
	if *singleFile != "" {
		if *parserName != "" {
			// Parse only single file
			fmt.Printf("Parsing single file with %s: %s\n", *parserName, *singleFile)
			err := proc.ProcessFileParseOnly(ctx, *singleFile, *parserName)
			if err != nil {
				log.Fatalf("Error parsing file: %v", err)
			}
			fmt.Println("Result saved to database.")
			return
		}
		// Full processing (currently unreachable due to log.Fatal above if parser not set)
		processSingle(ctx, proc, *singleFile)
		return
	}

	if *inputFile == "" {
		// Check for filenames from stdin or command line args
		if flag.NArg() > 0 {
			if *parserName != "" {
				proc.ParseBatch(ctx, flag.Args(), *parserName)
				return
			}
			processBatch(ctx, proc, cfg, flag.Args())
		} else {
			flag.Usage()
			fmt.Println("\nExamples:")
			fmt.Println("  comic-parser -parser regex -file \"Amazing Spider-Man 001 (2018).cbz\"")
			fmt.Println("  comic-parser -parser llm -input filenames.txt")
			fmt.Println("  comic-parser -generate-config")
			os.Exit(1)
		}
		return
	}

	// Load filenames from input file
	filenames, err := loadFilenames(*inputFile)
	if err != nil {
		log.Fatalf("Error loading input file: %v", err)
	}

	if len(filenames) == 0 {
		log.Fatal("No filenames to process")
	}

	fmt.Printf("Loaded %d filenames to process\n", len(filenames))

	if *parserName != "" {
		// Parse Only Mode
		fmt.Printf("Starting parse-only batch with parser: %s\n", *parserName)
		startTime := time.Now()
		proc.ParseBatch(ctx, filenames, *parserName)

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
		return
	}

	processBatch(ctx, proc, cfg, filenames)
}

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
