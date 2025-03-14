package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"strings"
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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip config check for config command itself
			if cmd.CommandPath() == "xander config" || 
			   strings.HasPrefix(cmd.CommandPath(), "xander config ") {
				return
			}
			
			// Check if any configuration is needed based on command
			switch cmd.Name() {
			case "comic":
				promptForComicVineConfig()
			case "fetch":
				promptForAPIConfig()
			// Add cases for other commands that need specific configs
			}
		},
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

// promptForComicVineConfig checks and prompts for ComicVine API key
func promptForComicVineConfig() {
	if cfg.ComicVineAPIKey == "" {
		fmt.Println("ComicVine API key is not configured. This is needed for comic metadata retrieval.")
		fmt.Println("You can get an API key from https://comicvine.gamespot.com/api/")
		
		if promptYesNo("Would you like to configure it now?") {
			apiKey := promptForInput("Enter your ComicVine API key")
			if apiKey != "" {
				cfg.ComicVineAPIKey = apiKey
				saveConfig()
			}
		}
	}
}

// promptForAPIConfig checks and prompts for general API key
func promptForAPIConfig() {
	if cfg.APIKey == "" {
		fmt.Println("API key is not configured. This is needed for the fetch command.")
		
		if promptYesNo("Would you like to configure it now?") {
			apiKey := promptForInput("Enter your API key")
			if apiKey != "" {
				cfg.APIKey = apiKey
				saveConfig()
			}
		}
	}
}

// promptForTVDBConfig checks and prompts for TVDB API key
func promptForTVDBConfig() {
	if cfg.TVDBAPIKey == "" {
		fmt.Println("TVDB API key is not configured. This is needed for TV show metadata retrieval.")
		fmt.Println("You can get an API key from https://thetvdb.com/api-information")
		
		if promptYesNo("Would you like to configure it now?") {
			apiKey := promptForInput("Enter your TVDB API key")
			if apiKey != "" {
				cfg.TVDBAPIKey = apiKey
				saveConfig()
			}
		}
	}
}

// Helper function to prompt for yes/no response
func promptYesNo(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/n): ", prompt)
	
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// Helper function to prompt for input
func promptForInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", prompt)
	
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(input)
}

// Helper function to save config
func saveConfig() {
	if err := cfg.Save(); err != nil {
		fmt.Printf("Failed to save configuration: %v\n", err)
	} else {
		fmt.Println("Configuration saved successfully")
	}
}

func init() {
	log.SetOutput(io.Discard)

	// dbPath is now set by the --save option in the comic command

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
