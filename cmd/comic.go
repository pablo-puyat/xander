package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"xander/internal/comic"
	"xander/internal/comicvine"
	"xander/internal/csv"
	"xander/internal/parse"
	"xander/internal/storage"
)

var (
	comicInputFile string
	comicOutputFormat string
	comicVerbose bool
	comicDbPath string // Path to save data, empty means use default location
	comicDryRun bool   // Dry run mode - parse only, don't query API
)

var comicCmd = &cobra.Command{
	Use:   "comicvine [string]",
	Short: "Get comic metadata for strings",
	Long: `Get metadata for string with comic-like filenames using ComicVine API.
string can be provided as arguments or read from a file using the --input flag.
strings should follow one of these formats:
  - "Series (Year) #Issue" 
  - "Publisher - Series (Year) #Issue"
  - "Series (Year) (digital) (Group)"
  - "Series 001 (Year) (digital) (Group)"
  - "Series - Title 000 (Year) (digital) (Group)"
  - "Series v01 - Title (Year) (digital) (Group)"
  - "Series 01 (of 08) (Year) (digital) (Group)"
  - "YYYY-MM - Title (digital) (Group)"
  - "YYYY (Year) (digital) (Group)"
  - "Series.Title.Month.Year.Format.Group"
  - "Series 001"
  
File extensions are ignored, so any string that follows the naming pattern can be processed.`,
	Run: runComicCmd,
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test the database fix",
	Long:  "Create test data in the database with unique volume information",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize storage with default path
		store, err := storage.GetStorage(storage.SQLite, "")
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			return
		}
		defer store.Close()
		
		// Create test comics with unique volume data
		testComics := []struct {
			filename string
			series   string
			issue    string
			year     string
			publisher string
			volumeID int
		}{
			{"Batman (2016) #001.cbz", "Batman", "001", "2016", "DC Comics", 101},
			{"DC Comics - The Flash (2016) #001.cbr", "The Flash", "001", "2016", "DC Comics", 102},
			{"Amazing Spider-Man Vol. 5 (2018) #001.cbz", "Amazing Spider-Man", "001", "2018", "Marvel", 103},
			{"Daredevil (2019) #001.cbz", "Daredevil", "001", "2019", "Marvel", 104},
		}
		
		for _, tc := range testComics {
			// Create volume JSON with unique ID
			volumeJSON := map[string]interface{}{
				"id": tc.volumeID,
				"name": tc.series,
				"issue_number": tc.issue,
				"api_detail_url": fmt.Sprintf("https://comicvine.gamespot.com/api/volume/4050-%d/", tc.volumeID),
				"site_detail_url": fmt.Sprintf("https://comicvine.gamespot.com/%s/4050-%d/", 
					strings.ToLower(strings.ReplaceAll(tc.series, " ", "-")), tc.volumeID),
			}
			
			// Create result
			result := &comicvine.Result{
				Filename:    tc.filename,
				Series:      tc.series,
				Issue:       tc.issue,
				Year:        tc.year,
				Publisher:   tc.publisher,
				ComicVineID: tc.volumeID + 1000, // Just make up a unique ID
				Title:       fmt.Sprintf("%s #%s", tc.series, tc.issue),
				Volume:      volumeJSON,
			}
			
			// Store in database
			if err := store.StoreComic(result); err != nil {
				fmt.Printf("Error storing comic '%s #%s' in database: %v\n", 
					tc.series, tc.issue, err)
				continue
			}
			
			fmt.Printf("Successfully added test data for %s #%s with unique volume ID %d\n", 
				tc.series, tc.issue, tc.volumeID)
		}
		
		// Show success message
		fmt.Println("Test data created. Run 'sqlite3 ~/.local/share/xander/xander.db \"SELECT id, series, issue, volume_json FROM comics;\"' to verify.")
	},
}

func init() {
	rootCmd.AddCommand(comicCmd)
	rootCmd.AddCommand(testCmd)
	
	comicCmd.Flags().StringVar(
		&comicInputFile,
		"input",
		"",
		"path to a file containing a list of strings (one per line)",
	)

	comicCmd.Flags().StringVar(
		&comicOutputFormat,
		"format",
		"text",
		"output format (text, json, or csv)",
	)
	
	comicCmd.Flags().BoolVar(
		&comicVerbose,
		"verbose",
		false,
		"enable verbose API logging",
	)
	
	// Mark the comicDbPath flag as having an optional value
	comicCmd.Flags().StringVar(
		&comicDbPath,
		"save",
		"",
		"store results in database (optionally specify custom database path)",
	)
	
	// Allow the --save flag to be used without a value
	comicCmd.Flags().Lookup("save").NoOptDefVal = "DEFAULT"
	
	// Add dry-run flag to test parsing without API queries
	comicCmd.Flags().BoolVar(
		&comicDryRun,
		"dry-run",
		false,
		"parse filenames only, don't query API or save to database",
	)
}

func runComicCmd(cmd *cobra.Command, args []string) {
	// Set up logging if verbose mode is enabled
	if comicVerbose {
		log.SetOutput(os.Stdout)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
		log.Println("Verbose mode enabled")
	}
	
	// Check if API key is configured
	if cfg.ComicVineAPIKey == "" {
		fmt.Println("ComicVine API key not set. Please configure it first.")
		return
	}

	var filenames []string

	// Get filenames from input file if provided
	if comicInputFile != "" {
		file, err := os.Open(comicInputFile)
		if err != nil {
			fmt.Printf("Error opening input file: %v\n", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			filename := strings.TrimSpace(scanner.Text())
			if filename != "" && !strings.HasPrefix(filename, "#") {
				filenames = append(filenames, filename)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading input file: %v\n", err)
			return
		}
	}

	// Add filenames from command line arguments
	filenames = append(filenames, args...)

	if len(filenames) == 0 {
		fmt.Println("No files provided. Please provide string to parse or use --input flag.")
		return
	}

	// Create the service
	service := comicvine.NewComicService(cfg.ComicVineAPIKey, comicVerbose)

	// Determine if we should check for existing entries in the database
	var store storage.Storage
	var shouldCheckDb bool
	
	// If the save flag is set, we can open the database early to check for existing entries
	if cmd.Flags().Changed("save") {
		// Determine the database path to use
		dbPath := comicDbPath
		
		// If --save was used without a value (or with the special DEFAULT value)
		if dbPath == "DEFAULT" {
			dbPath = "" // Empty string means use default location
		}
		
		// Initialize storage with specified path (empty means use default location)
		var err error
		store, err = storage.GetStorage(storage.SQLite, dbPath)
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			return
		}
		defer store.Close()
		
		shouldCheckDb = true
	}
	
	// Track files we can't parse or skip for various reasons
	type skippedFile struct {
		filename string
		reason   string
	}
	var skippedFiles []skippedFile
	
	// Pre-filter files - verify we can parse them before checking DB or API
	var validFilenames []string
	var parsedResults []struct {
		filename string
		series string
		issue string
		year string
		publisher string
	}
	
	for _, filename := range filenames {
		// Try to parse the filename first - only proceed with valid patterns
		series, issue, year, publisher, err := parse.ParseComicFilename(filename)
		if err != nil {
			// Can't parse the filename - doesn't match the expected patterns
			skippedFiles = append(skippedFiles, skippedFile{
				filename: filename,
				reason:   "Doesn't match any supported comic filename format",
			})
			continue
		}
		
		// Store parsed results for dry-run mode
		parsedResults = append(parsedResults, struct {
			filename string
			series string
			issue string
			year string
			publisher string
		}{
			filename: filename,
			series: series,
			issue: issue,
			year: year,
			publisher: publisher,
		})
		
		// Filename passed validation, add to list of files to process
		validFilenames = append(validFilenames, filename)
	}
	
	// If dry-run mode is enabled, show parsing results and exit
	if comicDryRun {
		fmt.Printf("Dry run results (parsed %d of %d files):\n\n", len(parsedResults), len(filenames))
		for _, pr := range parsedResults {
			fmt.Printf("Filename: %s\n", pr.filename)
			fmt.Printf("  Series: %s\n", pr.series)
			fmt.Printf("  Issue: %s\n", pr.issue)
			fmt.Printf("  Year: %s\n", pr.year)
			if pr.publisher != "" {
				fmt.Printf("  Publisher: %s\n", pr.publisher)
			}
			fmt.Println()
		}
		
		// Report skipped files in dry-run mode too
		if len(skippedFiles) > 0 {
			fmt.Printf("Skipped %d files due to parsing issues:\n", len(skippedFiles))
			for _, sf := range skippedFiles {
				fmt.Printf("  %s: %s\n", sf.filename, sf.reason)
			}
		}
		
		return
	}
	
	// Process each file - check if it exists in the database BEFORE sending to API
	var apiFilenames []string
	
	for _, filename := range validFilenames {
		// Check if the filename exists in the database
		if shouldCheckDb {
			// Use only the filename to check the database - before parsing
			exists, err := store.FilenameExistsInDb(filename)
			if err != nil {
				fmt.Printf("Error checking database for %s: %v\n", filename, err)
			}
			
			if exists {
				// Skip files already in the database
				fmt.Printf("Skipping %s (already in database)\n", filename)
			} else {
				// Not found, need to get from API
				apiFilenames = append(apiFilenames, filename)
			}
		} else {
			// Always get from API if not checking DB
			apiFilenames = append(apiFilenames, filename)
		}
	}
	
	// If saving to DB and no files to process, we're done
	if shouldCheckDb && len(apiFilenames) == 0 && len(skippedFiles) == 0 {
		fmt.Println("All files are already in the database - nothing to do.")
		return
	}
	
	// Report skipped files
	if len(skippedFiles) > 0 {
		fmt.Printf("\nSkipped %d files due to parsing issues:\n", len(skippedFiles))
		for _, sf := range skippedFiles {
			fmt.Printf("  %s: %s\n", sf.filename, sf.reason)
		}
		fmt.Println()
	}
	
	// Only call the API for valid files not found in the database
	if len(apiFilenames) == 0 {
		if len(skippedFiles) > 0 {
			fmt.Println("No valid files to process. Please check the skipped files list.")
		} else {
			fmt.Println("No files to process.")
		}
		return
	}
	
	apiResults, err := service.GetMetadataForFiles(apiFilenames)
	if err != nil {
		fmt.Printf("Error getting metadata from API: %v\n", err)
		return
	}
	
	// If we're saving to the database
	if cmd.Flags().Changed("save") {
		// Save API results to database
		if len(apiResults) > 0 {
			savedCount := 0
			total := len(apiResults)
			
			fmt.Println("\nSaving results to database...")
			
			for i, result := range apiResults {
				// Show progress before saving each comic
				fmt.Printf("Saving %d of %d: %s #%s\n", i+1, total, result.Series, result.Issue)
				
				if err := store.StoreComic(result); err != nil {
					fmt.Printf("Error storing comic '%s #%s' in database: %v\n", 
						result.Series, result.Issue, err)
					continue
				}
				savedCount++
			}
			
			// Show success message with database location
			dbLocation := dbPath
			if dbLocation == "" {
				// If using default location, show that to the user
				userHome, _ := os.UserHomeDir()
				if userHome != "" {
					dbLocation = filepath.Join(userHome, ".local", "share", "xander", "xander.db")
				} else {
					dbLocation = "default location"
				}
			}
			
			fmt.Printf("Saved %d comics to the database at: %s\n", savedCount, dbLocation)
		} else {
			fmt.Println("No comics were saved to the database.")
		}
	} else {
		// Only output results if we're not saving to the database
		if comicOutputFormat == "json" {
			// Convert API results to domain model
			comics := make([]*comic.Comic, len(apiResults))
			for i, result := range apiResults {
				comics[i] = result.ToComic()
			}
			outputJSON(comics)
		} else if comicOutputFormat == "csv" {
			// Convert API results to domain model
			comics := make([]*comic.Comic, len(apiResults))
			for i, result := range apiResults {
				comics[i] = result.ToComic()
			}
			outputCSV(comics)
		} else {
			outputText(apiResults)
		}
	}
}

func outputText(results []*comicvine.Result) {
	fmt.Printf("Found metadata for %d comics:\n\n", len(results))
	
	for _, result := range results {
		fmt.Printf("Original String: %s\n", result.Filename)
		fmt.Printf("Series: %s\n", result.Series)
		fmt.Printf("Issue: %s\n", result.Issue)
		fmt.Printf("Year: %s\n", result.Year)
		if result.Publisher != "" {
			fmt.Printf("Publisher: %s\n", result.Publisher)
		}
		fmt.Printf("ComicVine ID: %d\n", result.ComicVineID)
		fmt.Printf("Title: %s\n", result.Title)
		fmt.Printf("Cover URL: %s\n", result.CoverURL)
		if result.Description != "" {
			fmt.Printf("Description: %s\n", result.Description)
		}
		fmt.Println()
	}
}

func outputJSON(comics []*comic.Comic) {
	// Using the json package would be better, but for simplicity
	// we'll just manually construct a JSON string
	fmt.Println("[")
	for i, comic := range comics {
		fmt.Printf("  {\n")
		fmt.Printf("    \"original_string\": %q,\n", comic.Filename)
		fmt.Printf("    \"series\": %q,\n", comic.Series)
		fmt.Printf("    \"issue\": %q,\n", comic.Issue)
		fmt.Printf("    \"year\": %q,\n", comic.Year)
		fmt.Printf("    \"publisher\": %q,\n", comic.Publisher)
		fmt.Printf("    \"comicvine_id\": %d,\n", comic.ComicVineID)
		fmt.Printf("    \"title\": %q,\n", comic.Title)
		fmt.Printf("    \"cover_url\": %q,\n", comic.CoverURL)
		fmt.Printf("    \"description\": %q\n", comic.Description)
		fmt.Printf("  }")
		
		if i < len(comics)-1 {
			fmt.Println(",")
		} else {
			fmt.Println()
		}
	}
	fmt.Println("]")
}

func outputCSV(comics []*comic.Comic) {
	csvString, err := csv.ComicToCSV(comics)
	if err != nil {
		fmt.Printf("Error generating CSV: %v\n", err)
		return
	}
	fmt.Println(csvString)
}