package tui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"comic-parser/internal/comicvine"
	"comic-parser/internal/config"
	"comic-parser/internal/models"
	"comic-parser/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModel_Update_Search(t *testing.T) {
	// 1. Setup temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := storage.NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// 2. Insert test data
	item := &models.ParsedFilename{
		OriginalFilename: "Test.Comic.001.cbr",
		Title:            "Test Comic",
		IssueNumber:      "1",
		Confidence:       "high",
	}
	if err := store.SaveParsedFilename(context.Background(), item, "test"); err != nil {
		t.Fatalf("Failed to save parsed filename: %v", err)
	}

	// 3. Setup mock ComicVine server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with empty volume search results to trigger fallback issue search
		if strings.Contains(r.URL.Path, "/search/") && strings.Contains(r.URL.Query().Get("resources"), "volume") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"results": []}`))
			return
		}

		// Respond with issue search results
		if strings.Contains(r.URL.Path, "/search/") && strings.Contains(r.URL.Query().Get("resources"), "issue") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"results": [
					{
						"id": 1,
						"name": "Test Issue",
						"issue_number": "1",
						"cover_date": "2020-01-01",
						"volume": {"id": 100, "name": "Test Volume"}
					}
				]
			}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// 4. Setup ComicVine Client
	cfg := &config.Config{
		ComicVineAPIKey:     "test-key",
		ComicVineAPIBaseURL: server.URL,
	}
	cvClient := comicvine.NewClient(cfg, server.Client())

	// 5. Initialize Model
	model, err := NewModel(context.Background(), store, cvClient)
	if err != nil {
		t.Fatalf("NewModel failed: %v", err)
	}

	// Verify initial state
	if len(model.items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(model.items))
	}

	// 6. Simulate Search Key Press
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m := updatedModel.(Model)

	if !m.searching {
		t.Error("Expected searching to be true after 's' key")
	}

	// Execute the command to get the result
	if cmd != nil {
		msg := cmd() // This runs the search
		// Feed the result back into Update
		finalModel, _ := m.Update(msg)
		fm := finalModel.(Model)

		if fm.searching {
			t.Error("Expected searching to be false after search completion")
		}
		if fm.searchErr != nil {
			t.Errorf("Unexpected search error: %v", fm.searchErr)
		}
		if len(fm.searchResults) == 0 {
			t.Error("Expected search results, got none")
		} else if fm.searchResults[0].ID != 1 {
			t.Errorf("Expected result ID 1, got %d", fm.searchResults[0].ID)
		}
	} else {
		t.Error("Expected a command after 's' key")
	}
}

func TestModel_Navigate(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := storage.NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Add two items
	store.SaveParsedFilename(context.Background(), &models.ParsedFilename{OriginalFilename: "1.cbr", Title: "1"}, "test")
	store.SaveParsedFilename(context.Background(), &models.ParsedFilename{OriginalFilename: "2.cbr", Title: "2"}, "test")

	model, _ := NewModel(context.Background(), store, nil)

	// Check initial index
	if model.index != 0 {
		t.Errorf("Expected index 0, got %d", model.index)
	}

	// Next
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m := updatedModel.(Model)
	if m.index != 1 {
		t.Errorf("Expected index 1, got %d", m.index)
	}

	// Next again (should stay at 1)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updatedModel.(Model)
	if m.index != 1 {
		t.Errorf("Expected index 1, got %d", m.index)
	}

	// Prev
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updatedModel.(Model)
	if m.index != 0 {
		t.Errorf("Expected index 0, got %d", m.index)
	}
}

func TestModel_View(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := storage.NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	store.SaveParsedFilename(context.Background(), &models.ParsedFilename{
		OriginalFilename: "ViewTest.cbr",
		Title:            "View Test",
		IssueNumber:      "1",
		Confidence:       "1.0",
	}, "test")

	model, _ := NewModel(context.Background(), store, nil)
	view := model.View()

	if !strings.Contains(view, "ViewTest.cbr") {
		t.Error("View output missing filename")
	}
	if !strings.Contains(view, "View Test") {
		t.Error("View output missing title")
	}
}
