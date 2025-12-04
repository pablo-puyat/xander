package main

import (
	"log"

	"comic-parser/internal/config"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dbPath  string
	verbose bool
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "comic-parser",
	Short: "A tool to parse comic book filenames and match them to ComicVine",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip config loading for config generation command
		if cmd.Name() == "init" && cmd.Parent().Name() == "config" {
			return
		}

		// Load configuration
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			// If config file is missing, we might want to proceed with defaults
			// unless the user specifically pointed to a file that doesn't exist
			if cfgFile != "config.json" {
				log.Fatalf("Error loading config: %v", err)
			}
			log.Fatalf("Error loading config: %v. Run 'comic-parser config init' to generate one.", err)
		}
		cfg.LoadFromEnv()

		// Override config with persistent flags
		if verbose {
			cfg.Verbose = true
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.json", "Path to configuration file")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "comics.db", "Database path for storing results")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
}
