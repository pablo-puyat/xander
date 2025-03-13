package cmd

import (
	"fmt"
	"log"
	"xander/internal/comicvine"

	"github.com/spf13/cobra"
)

var fetchCommand = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch comic information",
	Long: `Fetch comic information from ComicVine based on filename.
Usage: xander fetch`,
	Args:                  cobra.ExactArgs(0),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := fetch(); err != nil {
			log.Fatalf("Scan failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(fetchCommand)
}

func fetch() error {
	if cfg.APIKey == "" {
		return fmt.Errorf("API key not configured. Run 'xander config'")
	}

	// Use global verbose flag for consistency
	client := comicvine.NewClient(cfg.APIKey, false)

	// TODO: Parse filename to get series and issue number
	series := "Batman" // This will come from filename parsing
	issueNumber := "1" // This will come from filename parsing

	issue, err := client.GetIssue(series, issueNumber)
	if err != nil {
		return err
	}

	fmt.Printf("Found: %s #%s\n", issue.Volume.Name, issue.IssueNumber)
	return nil
}
