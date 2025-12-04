package main

import (
	"fmt"
	"log"

	"comic-parser/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a sample config file",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.DefaultConfig()
		cfg.AnthropicAPIKey = "your-anthropic-api-key-here"
		cfg.ComicVineAPIKey = "your-comicvine-api-key-here"
		if err := cfg.SaveConfig("config.sample.json"); err != nil {
			log.Fatalf("Error generating config: %v", err)
		}
		fmt.Println("Generated config.sample.json - copy to config.json and add your API keys")
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)
}
