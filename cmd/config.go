package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setAPIKeyCmd)
	configCmd.AddCommand(getAPIKeyCmd)
	configCmd.AddCommand(setComicVineKeyCmd)
	configCmd.AddCommand(getComicVineKeyCmd)
	configCmd.AddCommand(setTVDBKeyCmd)
	configCmd.AddCommand(getTVDBKeyCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Configure API keys and settings for various services",
	Run: func(cmd *cobra.Command, args []string) {
		// Display current configuration
		fmt.Println("Current configuration:")
		
		if cfg.APIKey != "" {
			fmt.Println("API key: [CONFIGURED]")
		} else {
			fmt.Println("API key: [NOT CONFIGURED]")
		}
		
		if cfg.ComicVineAPIKey != "" {
			fmt.Println("ComicVine API key: [CONFIGURED]")
		} else {
			fmt.Println("ComicVine API key: [NOT CONFIGURED]")
		}
		
		if cfg.TVDBAPIKey != "" {
			fmt.Println("TVDB API key: [CONFIGURED]")
		} else {
			fmt.Println("TVDB API key: [NOT CONFIGURED]")
		}
	},
}

var setAPIKeyCmd = &cobra.Command{
	Use:   "set-key [apiKey]",
	Short: "Set the API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.APIKey = args[0]
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("API key updated successfully")
		return nil
	},
}

var getAPIKeyCmd = &cobra.Command{
	Use:   "get-key",
	Short: "Get the current API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.APIKey == "" {
			return fmt.Errorf("no API key configured")
		}
		fmt.Printf("Current API key: %s\n", cfg.APIKey)
		return nil
	},
}

var setComicVineKeyCmd = &cobra.Command{
	Use:   "set-comicvine-key [apiKey]",
	Short: "Set the ComicVine API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.ComicVineAPIKey = args[0]
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("ComicVine API key updated successfully")
		return nil
	},
}

var getComicVineKeyCmd = &cobra.Command{
	Use:   "get-comicvine-key",
	Short: "Get the current ComicVine API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.ComicVineAPIKey == "" {
			return fmt.Errorf("no ComicVine API key configured")
		}
		fmt.Printf("Current ComicVine API key: %s\n", cfg.ComicVineAPIKey)
		return nil
	},
}

var setTVDBKeyCmd = &cobra.Command{
	Use:   "set-tvdb-key [apiKey]",
	Short: "Set the TVDB API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg.TVDBAPIKey = args[0]
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("TVDB API key updated successfully")
		return nil
	},
}

var getTVDBKeyCmd = &cobra.Command{
	Use:   "get-tvdb-key",
	Short: "Get the current TVDB API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.TVDBAPIKey == "" {
			return fmt.Errorf("no TVDB API key configured")
		}
		fmt.Printf("Current TVDB API key: %s\n", cfg.TVDBAPIKey)
		return nil
	},
}
