package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"xander/internal/comicvine"
	"xander/internal/storage"

	"github.com/spf13/cobra"
)

var (
	dbExportFile string
	dbImportFile string
	dbFormat     string
	dbQuery      string
	dbLimit      int
	dbOffset     int
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database operations",
	Long:  `Commands for interacting with the comic metadata database.`,
}

var dbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List comics in the database",
	Long:  `List comics stored in the database with optional filtering.`,
	Run:   runDbListCmd,
}

var dbExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export database to file",
	Long:  `Export comics from the database to a JSON file.`,
	Run:   runDbExportCmd,
}

var dbImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import metadata from file",
	Long:  `Import comic metadata from a JSON file into the database.`,
	Run:   runDbImportCmd,
}

var dbStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show database statistics",
	Long:  `Display statistics about the comics stored in the database.`,
	Run:   runDbStatsCmd,
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbListCmd, dbExportCmd, dbImportCmd, dbStatsCmd)

	// Flags for list command
	dbListCmd.Flags().StringVar(&dbQuery, "query", "", "Filter by series name (case-insensitive)")
	dbListCmd.Flags().IntVar(&dbLimit, "limit", 20, "Maximum number of results to return")
	dbListCmd.Flags().IntVar(&dbOffset, "offset", 0, "Number of results to skip")
	dbListCmd.Flags().StringVar(&dbFormat, "format", "text", "Output format (text or json)")

	// Flags for export command
	dbExportCmd.Flags().StringVar(&dbExportFile, "output", "comics_export.json", "Output file path")
	dbExportCmd.Flags().StringVar(&dbQuery, "query", "", "Filter by series name (case-insensitive)")

	// Flags for import command
	dbImportCmd.Flags().StringVar(&dbImportFile, "input", "", "Input JSON file path")
	dbImportCmd.MarkFlagRequired("input")
}

func runDbListCmd(cmd *cobra.Command, args []string) {
	// Initialize storage
	store, err := storage.GetStorage(storage.SQLite, dbPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	defer store.Close()

	// Create filter
	filter := storage.NewFilter()
	filter.Series = dbQuery
	filter.Limit = dbLimit
	filter.Offset = dbOffset

	// Get comics
	comics, err := store.GetComicsByFilter(filter)
	if err != nil {
		fmt.Printf("Error querying database: %v\n", err)
		return
	}

	// Display results
	if len(comics) == 0 {
		fmt.Println("No comics found.")
		return
	}

	if dbFormat == "json" {
		// Output as JSON
		jsonData, err := json.MarshalIndent(comics, "", "  ")
		if err != nil {
			fmt.Printf("Error encoding to JSON: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
	} else {
		// Output as text
		fmt.Printf("Found %d comics:\n\n", len(comics))
		for _, comic := range comics {
			fmt.Printf("Series: %s\n", comic.Series)
			fmt.Printf("Issue: %s\n", comic.Issue)
			if comic.Year != "" {
				fmt.Printf("Year: %s\n", comic.Year)
			}
			if comic.Publisher != "" {
				fmt.Printf("Publisher: %s\n", comic.Publisher)
			}
			fmt.Printf("Title: %s\n", comic.Title)
			fmt.Printf("ComicVine ID: %d\n", comic.ComicVineID)
			fmt.Println()
		}
		
		// Show pagination info
		if dbLimit <= len(comics) {
			fmt.Printf("Showing %d of %d+ results (use --offset %d to see more)\n", 
				len(comics), dbOffset+len(comics), dbOffset+dbLimit)
		} else {
			fmt.Printf("Showing %d results\n", len(comics))
		}
	}
}

func runDbExportCmd(cmd *cobra.Command, args []string) {
	// Initialize storage
	store, err := storage.GetStorage(storage.SQLite, dbPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	defer store.Close()

	// Create filter
	filter := storage.NewFilter()
	filter.Series = dbQuery
	filter.Limit = 10000 // Large limit to get all records

	// Get comics
	comics, err := store.GetComicsByFilter(filter)
	if err != nil {
		fmt.Printf("Error querying database: %v\n", err)
		return
	}

	if len(comics) == 0 {
		fmt.Println("No comics found to export.")
		return
	}

	// Export to JSON file
	jsonData, err := json.MarshalIndent(comics, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding to JSON: %v\n", err)
		return
	}

	err = os.WriteFile(dbExportFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}

	fmt.Printf("Exported %d comics to %s\n", len(comics), dbExportFile)
}

func runDbImportCmd(cmd *cobra.Command, args []string) {
	// Read the input file
	jsonData, err := os.ReadFile(dbImportFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Parse the JSON
	var comics []*comicvine.Result
	err = json.Unmarshal(jsonData, &comics)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	if len(comics) == 0 {
		fmt.Println("No comics found in the import file.")
		return
	}

	// Initialize storage
	store, err := storage.GetStorage(storage.SQLite, dbPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	defer store.Close()

	// Import comics
	importCount := 0
	for _, comic := range comics {
		err = store.StoreComic(comic)
		if err != nil {
			fmt.Printf("Error storing comic '%s #%s': %v\n", comic.Series, comic.Issue, err)
			continue
		}
		importCount++
	}

	fmt.Printf("Successfully imported %d of %d comics\n", importCount, len(comics))
}

func runDbStatsCmd(cmd *cobra.Command, args []string) {
	// Initialize storage
	store, err := storage.GetStorage(storage.SQLite, dbPath)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}
	defer store.Close()

	// Get all comics
	comics, err := store.GetComics()
	if err != nil {
		fmt.Printf("Error querying database: %v\n", err)
		return
	}

	if len(comics) == 0 {
		fmt.Println("No comics in the database.")
		return
	}

	// Calculate statistics
	seriesMap := make(map[string]int)
	publisherMap := make(map[string]int)
	yearMap := make(map[string]int)
	
	for _, comic := range comics {
		seriesMap[comic.Series]++
		if comic.Publisher != "" {
			publisherMap[comic.Publisher]++
		}
		if comic.Year != "" {
			yearMap[comic.Year]++
		}
	}

	// Display statistics
	fmt.Printf("Database Statistics\n")
	fmt.Printf("------------------\n")
	fmt.Printf("Total Comics: %d\n", len(comics))
	fmt.Printf("Unique Series: %d\n", len(seriesMap))
	fmt.Printf("Unique Publishers: %d\n", len(publisherMap))
	fmt.Printf("Year Range: %d years\n", len(yearMap))
	
	// Show most common series
	fmt.Printf("\nTop Series:\n")
	seriesCounts := make(map[string]int)
	for _, comic := range comics {
		seriesCounts[comic.Series]++
	}
	
	// Convert to a sortable slice
	type seriesCount struct {
		name  string
		count int
	}
	var seriesList []seriesCount
	for name, count := range seriesCounts {
		seriesList = append(seriesList, seriesCount{name, count})
	}
	
	// Sort by count (descending)
	sort.Slice(seriesList, func(i, j int) bool {
		return seriesList[i].count > seriesList[j].count
	})
	
	// Display top 10 or fewer
	limit := 10
	if len(seriesList) < limit {
		limit = len(seriesList)
	}
	
	for i := 0; i < limit; i++ {
		fmt.Printf("  %s: %d issues\n", seriesList[i].name, seriesList[i].count)
	}
}