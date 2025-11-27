package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"comic-parser/models"
)

func TestStorage(t *testing.T) {
	dbPath := "test_comics.db"
	defer os.Remove(dbPath)

	store, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Test case 1: Basic result with match
	result := &models.ProcessingResult{
		Filename:         "Amazing Spider-Man 001.cbz",
		Success:          true,
		ProcessedAt:      time.Now(),
		ProcessingTimeMS: 1234,
		Match: &models.MatchResult{
			ParsedInfo: models.ParsedFilename{
				OriginalFilename: "Amazing Spider-Man 001.cbz",
				Title:            "Amazing Spider-Man",
				IssueNumber:      "001",
				Confidence:       "high",
			},
			MatchConfidence: "high",
			Reasoning:       "Exact match",
			SelectedIssue: &models.ComicVineIssue{
				ID:          12345,
				Name:        "Spider-Man vs. Goblin",
				IssueNumber: "1",
				Volume: models.VolumeRef{
					ID:        4050,
					Name:      "Amazing Spider-Man",
					SiteURL:   "http://example.com/vol",
					Publisher: "Marvel",
				},
				SiteDetailURL: "http://example.com/issue",
				Image: models.ImageRef{
					SmallURL: "http://example.com/small.jpg",
				},
			},
		},
	}

	ctx := context.Background()
	if err := store.SaveResult(ctx, result); err != nil {
		t.Fatalf("Failed to save result: %v", err)
	}

	// Verify data in DB
	var count int
	err = store.db.QueryRow("SELECT count(*) FROM processing_results").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query processing_results: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 processing_result, got %d", count)
	}

	err = store.db.QueryRow("SELECT count(*) FROM parsed_filenames").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query parsed_filenames: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 parsed_filename, got %d", count)
	}

	// Check volume and issue
	err = store.db.QueryRow("SELECT count(*) FROM comic_vine_volumes").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query volumes: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 volume, got %d", count)
	}

	err = store.db.QueryRow("SELECT count(*) FROM comic_vine_issues").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query issues: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 issue, got %d", count)
	}

	// Test case 2: Update existing result
	result.ProcessingTimeMS = 5000
	if err := store.SaveResult(ctx, result); err != nil {
		t.Fatalf("Failed to update result: %v", err)
	}

	// Should still be 1 row
	err = store.db.QueryRow("SELECT count(*) FROM processing_results").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query processing_results: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 processing_result after update, got %d", count)
	}

	var timeMs int64
	err = store.db.QueryRow("SELECT processing_time_ms FROM processing_results WHERE filename = ?", result.Filename).Scan(&timeMs)
	if err != nil {
		t.Fatalf("Failed to query time: %v", err)
	}
	if timeMs != 5000 {
		t.Errorf("Expected time 5000, got %d", timeMs)
	}
}
