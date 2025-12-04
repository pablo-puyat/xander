package storage

import (
	"context"
	"os"
	"testing"

	"comic-parser/internal/models"
)

func TestListParsedFilenames(t *testing.T) {
	dbPath := "test_comics_list.db"
	defer os.Remove(dbPath)

	store, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// 1. Insert some parsed filenames
	p1 := &models.ParsedFilename{
		OriginalFilename: "file1.cbz",
		Title:            "Title 1",
		IssueNumber:      "1",
		Confidence:       "high",
		Notes:            "note 1",
	}
	if err := store.SaveParsedFilename(ctx, p1, "regex"); err != nil {
		t.Fatalf("Failed to save p1: %v", err)
	}

	p2 := &models.ParsedFilename{
		OriginalFilename: "file2.cbz",
		Title:            "Title 2",
		IssueNumber:      "2",
		Confidence:       "medium",
	}
	if err := store.SaveParsedFilename(ctx, p2, "llm"); err != nil {
		t.Fatalf("Failed to save p2: %v", err)
	}

	// 2. List them
	items, err := store.ListParsedFilenames(ctx)
	if err != nil {
		t.Fatalf("Failed to list: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	// Check order (DESC by ID, so p2 then p1)
	if items[0].OriginalFilename != "file2.cbz" {
		t.Errorf("Expected first item to be file2.cbz, got %s", items[0].OriginalFilename)
	}
	if items[1].OriginalFilename != "file1.cbz" {
		t.Errorf("Expected second item to be file1.cbz, got %s", items[1].OriginalFilename)
	}

	// Check fields
	if items[0].Notes != "" { // p2 has no notes
		t.Errorf("Expected p2 notes empty, got %s", items[0].Notes)
	}
	if items[1].Notes != "note 1" {
		t.Errorf("Expected p1 notes 'note 1', got %s", items[1].Notes)
	}
}
