package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"comic-parser/internal/comicvine"
	"comic-parser/internal/storage"
	"comic-parser/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch TUI to view parsed results",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize Storage
		store, err := storage.NewStorage(dbPath)
		if err != nil {
			log.Fatalf("Error initializing storage: %v", err)
		}
		defer store.Close()

		// Create shared HTTP client
		httpClient := &http.Client{
			Timeout: 60 * time.Second,
		}

		cvClient := comicvine.NewClient(cfg, httpClient)

		// Create context
		ctx := context.Background()

		model, err := tui.NewModel(ctx, store, cvClient)
		if err != nil {
			log.Fatalf("Error initializing TUI: %v", err)
		}

		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			log.Fatalf("Error running TUI: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
