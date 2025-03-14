package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"xander/internal/comic"
	"xander/internal/comicvine"
	"xander/internal/csv"
)

var (
	comicInputFile string
	comicOutputFormat string
	comicVerbose bool
)

var comicCmd = &cobra.Command{
	Use:   "comic [filenames]",
	Short: "Get comic metadata for files",
	Long: `Get metadata for files with comic-like filenames using ComicVine API.
Files can be provided as arguments or read from a file using the --input flag.
Filenames should follow one of these formats:
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
  
File extensions are ignored, so any file that follows the naming pattern can be processed.`,
	Run: runComicCmd,
}

func init() {
	rootCmd.AddCommand(comicCmd)

	comicCmd.Flags().StringVar(
		&comicInputFile,
		"input",
		"",
		"path to a file containing a list of comic files (one per line)",
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
		fmt.Println("No files provided. Please provide files as arguments or use --input flag.")
		return
	}

	// Create the service
	service := comicvine.NewComicService(cfg.ComicVineAPIKey, comicVerbose)

	// Get metadata
	results, err := service.GetMetadataForFiles(filenames)
	if err != nil {
		fmt.Printf("Error getting metadata: %v\n", err)
		return
	}

	// Output results
	if comicOutputFormat == "json" {
		// Convert API results to domain model
		comics := make([]*comic.Comic, len(results))
		for i, result := range results {
			comics[i] = result.ToComic()
		}
		outputJSON(comics)
	} else if comicOutputFormat == "csv" {
		// Convert API results to domain model
		comics := make([]*comic.Comic, len(results))
		for i, result := range results {
			comics[i] = result.ToComic()
		}
		outputCSV(comics)
	} else {
		outputText(results)
	}
}

func outputText(results []*comicvine.Result) {
	fmt.Printf("Found metadata for %d comics:\n\n", len(results))
	
	for _, result := range results {
		fmt.Printf("File: %s\n", result.Filename)
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
		fmt.Printf("    \"filename\": %q,\n", comic.Filename)
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
