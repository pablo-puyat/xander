package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"comic-parser/internal/comicvine"
	"comic-parser/internal/llm"
	"comic-parser/internal/parser"
	"comic-parser/internal/processor"
	"comic-parser/internal/selector"
	"comic-parser/internal/storage"

	"github.com/spf13/cobra"
)

var (
	inputFile    string
	outputFile   string
	outputFormat string
	workers      int
	interactive  bool
	singleFile   string
	parserName   string
)

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse comic book filenames",
	Long: `Parse comic book filenames from a file or arguments.
Uses either regex or LLM based parsing to extract information and match with ComicVine.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Override config with flags
		if workers > 0 {
			cfg.WorkerCount = workers
		}
		if outputFile != "" {
			cfg.OutputFile = outputFile
		}
		if outputFormat != "" {
			cfg.OutputFormat = outputFormat
		}

		if cmd.Flags().Changed("interactive") {
			cfg.Interactive = interactive
		}

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
		if parserName != "" {
			switch parserName {
			case "regex":
				p = parser.NewRegexParser()
			case "llm":
				p = parser.NewLLMParser(llmClient, cfg.RetryAttempts, cfg.RetryDelaySeconds)
			default:
				log.Fatalf("Unknown parser: %s (must be regex or llm)", parserName)
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

		// Initialize Storage
		var store *storage.Storage
		var err error
		store, err = storage.NewStorage(dbPath)
		if err != nil {
			log.Fatalf("Error initializing storage: %v", err)
		}
		defer store.Close()

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

		// Process single file
		if singleFile != "" {
			if parserName != "" {
				fmt.Printf("Parsing single file with %s: %s\n", parserName, singleFile)
				err := proc.ProcessFileParseOnly(ctx, singleFile, parserName)
				if err != nil {
					log.Fatalf("Error parsing file: %v", err)
				}
				fmt.Println("Result saved to database.")
				return
			}
			processSingle(ctx, proc, singleFile)
			return
		}

		var filenames []string
		if inputFile == "" {
			if len(args) > 0 {
				filenames = args
			} else {
				cmd.Usage()
				os.Exit(1)
			}
		} else {
			var err error
			filenames, err = loadFilenames(inputFile)
			if err != nil {
				log.Fatalf("Error loading input file: %v", err)
			}
		}

		if len(filenames) == 0 {
			log.Fatal("No filenames to process")
		}

		fmt.Printf("Loaded %d filenames to process\n", len(filenames))

		if parserName != "" {
			// Parse Only Mode
			fmt.Printf("Starting parse-only batch with parser: %s\n", parserName)
			startTime := time.Now()
			proc.ParseBatch(ctx, filenames, parserName)

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
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)

	parseCmd.Flags().StringVar(&inputFile, "input", "", "Input file containing filenames (one per line)")
	parseCmd.Flags().StringVar(&outputFile, "output", "results.json", "Output file for results")
	parseCmd.Flags().StringVar(&outputFormat, "format", "json", "Output format: json, csv, or sqlite")
	parseCmd.Flags().IntVar(&workers, "workers", 3, "Number of concurrent workers")
	parseCmd.Flags().BoolVar(&interactive, "interactive", false, "Enable interactive TUI mode")
	parseCmd.Flags().StringVar(&singleFile, "file", "", "Process a single filename (for testing)")
	parseCmd.Flags().StringVar(&parserName, "parser", "", "Parser to use: regex or llm (enables parse-only mode)")
}
