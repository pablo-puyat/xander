package selector

import (
	"context"
	"os"
	"testing"
	"time"

	"comic-parser/internal/models"
)

func TestTUISelector_Select(t *testing.T) {
	// Create a pipe to mock stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Save original stdin
	oldStdin := os.Stdin
	defer func() {
		os.Stdin = oldStdin
		w.Close()
	}()

	// Set stdin to our pipe
	os.Stdin = r

	// Simulate user input
	// "1\n" selects the first item
	go func() {
		time.Sleep(100 * time.Millisecond) // Give time for output to print
		w.Write([]byte("1\n"))
	}()

	parsed := &models.ParsedFilename{
		OriginalFilename: "Test Comic 001.cbz",
		Title:            "Test Comic",
		IssueNumber:      "1",
	}

	issues := []models.ComicVineIssue{
		{
			ID:          123,
			Name:        "Test Comic",
			IssueNumber: "1",
			Volume: models.VolumeRef{
				Name: "Test Comic Vol 1",
			},
		},
		{
			ID:          456,
			Name:        "Test Comic",
			IssueNumber: "2",
			Volume: models.VolumeRef{
				Name: "Test Comic Vol 1",
			},
		},
	}

	s := NewTUISelector()
	ctx := context.Background()

	result, err := s.Select(ctx, parsed, issues)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if result.SelectedIssue == nil {
		t.Fatal("Expected a selected issue, got nil")
	}

	if result.SelectedIssue.ID != 123 {
		t.Errorf("Expected issue ID 123, got %d", result.SelectedIssue.ID)
	}

	if result.MatchConfidence != "high" {
		t.Errorf("Expected high confidence, got %s", result.MatchConfidence)
	}
}

func TestTUISelector_NoMatch(t *testing.T) {
	// Create a pipe to mock stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Save original stdin
	oldStdin := os.Stdin
	defer func() {
		os.Stdin = oldStdin
		w.Close()
	}()

	// Set stdin to our pipe
	os.Stdin = r

	// Simulate user input
	// "0\n" selects No Match
	go func() {
		time.Sleep(100 * time.Millisecond)
		w.Write([]byte("0\n"))
	}()

	parsed := &models.ParsedFilename{
		OriginalFilename: "Test Comic 001.cbz",
		Title:            "Test Comic",
		IssueNumber:      "1",
	}

	issues := []models.ComicVineIssue{
		{ID: 123},
	}

	s := NewTUISelector()
	ctx := context.Background()

	result, err := s.Select(ctx, parsed, issues)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if result.SelectedIssue != nil {
		t.Errorf("Expected no selected issue, got %v", result.SelectedIssue)
	}

	if result.MatchConfidence != "none" {
		t.Errorf("Expected none confidence, got %s", result.MatchConfidence)
	}
}
