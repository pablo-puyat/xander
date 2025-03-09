package cmd

import (
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"xander/internal/config"
)

var (
	dbPath  string
	verbose bool
	logFile string

	rootCmd = &cobra.Command{
		Use:   "xander",
		Short: "Xander is a tool to investigate files.",
		Long: `Xander is a CLI / TUI to retrieve metadata for media files.  It can get meta data for cbrz files and tv shows.
Usage: 
- xander 
`,
	}
)

var (
	cfg *config.Config
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	log.SetOutput(io.Discard)

	rootCmd.PersistentFlags().StringVar(
		&dbPath,
		"database",
		"",
		"path to SQLite database file",
	)

	rootCmd.PersistentFlags().BoolVar(
		&verbose,
		"verbose",
		false,
		"enable verbose logging to stdout",
	)

	rootCmd.PersistentFlags().StringVar(
		&logFile,
		"log",
		"",
		"path to log file",
	)

	cobra.OnInitialize(func() {
		if logFile != "" {
			file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			log.SetOutput(file)
		} else if verbose {
			log.SetOutput(os.Stdout)
		} else {
			log.SetOutput(io.Discard)
		}

		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			// Initialize empty config if none exists
			cfg = config.NewConfig()
		}
	})

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
