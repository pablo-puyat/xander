package csv

import (
	"bytes"
	"strings"
	"testing"
	"xander/internal/comic"
)

func TestComicToCSV(t *testing.T) {
	// Create a sample comic result
	comicA := &comic.Comic{
		Filename:    "Batman (2016) #001.cbz",
		Series:      "Batman",
		Issue:       "001",
		Year:        "2016",
		Publisher:   "DC Comics",
		ComicVineID: 12345,
		Title:       "I Am Gotham, Part One",
		CoverURL:    "http://example.com/cover.jpg",
		Description: "A new era for the Dark Knight begins here!",
	}

	// Test single comic conversion
	csv, err := ComicToCSV([]*comic.Comic{comicA})
	if err != nil {
		t.Errorf("ComicToCSV returned an error: %v", err)
	}

	// CSV should have a header and one data row
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in CSV (header + data), got %d", len(lines))
	}

	// Check header contains expected fields
	expectedHeaders := []string{"Filename", "Series", "Issue", "Year", "Publisher", "ComicVineID", "Title"}
	for _, header := range expectedHeaders {
		if !strings.Contains(lines[0], header) {
			t.Errorf("Header missing expected field: %s", header)
		}
	}
	
	// Check data row contains comic values
	dataRow := lines[1]
	expectedValues := []string{"Batman (2016) #001.cbz", "Batman", "001", "2016", "DC Comics", "12345", "I Am Gotham, Part One"}
	for _, value := range expectedValues {
		if !strings.Contains(dataRow, value) {
			t.Errorf("Data row missing expected value: %s", value)
		}
	}
	
}

func TestWriteCSV(t *testing.T) {
	// Create sample comic results
	comics := []*comic.Comic{
		{
			Filename:    "Batman (2016) #001.cbz",
			Series:      "Batman",
			Issue:       "001",
			Year:        "2016",
			Publisher:   "DC Comics",
			ComicVineID: 12345,
			Title:       "I Am Gotham, Part One",
		},
		{
			Filename:    "The Flash (2016) #001.cbr",
			Series:      "The Flash",
			Issue:       "001",
			Year:        "2016",
			Publisher:   "DC Comics",
			ComicVineID: 67890,
			Title:       "Lightning Strikes Twice, Part One",
		},
	}

	// Create a buffer to write to
	var buf bytes.Buffer

	// Call WriteCSV
	err := WriteCSV(&buf, comics)
	if err != nil {
		t.Errorf("WriteCSV returned an error: %v", err)
	}

	// Check that the CSV was written correctly
	csv := buf.String()
	lines := strings.Split(strings.TrimSpace(csv), "\n")

	// Should have 3 lines (header + 2 data rows)
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines in CSV (header + 2 data), got %d", len(lines))
	}

	// Check for Batman data in the output
	if !strings.Contains(csv, "Batman") || !strings.Contains(csv, "12345") {
		t.Errorf("CSV missing Batman data")
	}

	// Check for Flash data in the output
	if !strings.Contains(csv, "The Flash") || !strings.Contains(csv, "67890") {
		t.Errorf("CSV missing Flash data")
	}
}

func TestCSVStorage(t *testing.T) {
	// Create a test CSV storage
	tempFile := t.TempDir() + "/test_comics.csv"
	storage, err := NewCSVStorage(tempFile)
	if err != nil {
		t.Fatalf("Failed to create CSV storage: %v", err)
	}
	defer storage.Close()

	// Test storing a comic
	comic := &comic.Comic{
		Filename:    "Batman (2016) #001.cbz",
		Series:      "Batman",
		Issue:       "001",
		Year:        "2016",
		Publisher:   "DC Comics",
		ComicVineID: 12345,
		Title:       "I Am Gotham, Part One",
	}

	err = storage.StoreComic(comic)
	if err != nil {
		t.Errorf("StoreComic returned an error: %v", err)
	}

	// Test retrieving all comics
	comics, err := storage.GetComics()
	if err != nil {
		t.Errorf("GetComics returned an error: %v", err)
	}

	if len(comics) != 1 {
		t.Errorf("Expected 1 comic, got %d", len(comics))
	}

	if comics[0].Series != "Batman" || comics[0].Issue != "001" {
		t.Errorf("Retrieved comic data doesn't match: got %v", comics[0])
	}
}

func TestCSVStorageWithFilter(t *testing.T) {
	// Create a test CSV storage with multiple comics
	tempFile := t.TempDir() + "/test_comics_filter.csv"
	storage, err := NewCSVStorage(tempFile)
	if err != nil {
		t.Fatalf("Failed to create CSV storage: %v", err)
	}
	defer storage.Close()

	// Store multiple comics
	comics := []*comic.Comic{
		{
			Filename:    "Batman (2016) #001.cbz",
			Series:      "Batman",
			Issue:       "001",
			Year:        "2016",
			Publisher:   "DC Comics",
			ComicVineID: 12345,
		},
		{
			Filename:    "Batman (2016) #002.cbz",
			Series:      "Batman",
			Issue:       "002",
			Year:        "2016",
			Publisher:   "DC Comics",
			ComicVineID: 12346,
		},
		{
			Filename:    "The Flash (2016) #001.cbr",
			Series:      "The Flash",
			Issue:       "001",
			Year:        "2016",
			Publisher:   "DC Comics",
			ComicVineID: 67890,
		},
	}

	for _, c := range comics {
		err = storage.StoreComic(c)
		if err != nil {
			t.Errorf("StoreComic returned an error: %v", err)
		}
	}

	// Test filtering by series
	filter := CSVFilter{
		Series: "Batman",
	}
	
	filtered, err := storage.GetComicsByFilter(filter)
	if err != nil {
		t.Errorf("GetComicsByFilter returned an error: %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("Expected 2 Batman comics, got %d", len(filtered))
	}

	// Test filtering by issue
	filter = CSVFilter{
		Issue: "001",
	}
	
	filtered, err = storage.GetComicsByFilter(filter)
	if err != nil {
		t.Errorf("GetComicsByFilter returned an error: %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("Expected 2 comics with issue 001, got %d", len(filtered))
	}

	// Test filtering by series and issue
	filter = CSVFilter{
		Series: "Batman",
		Issue:  "001",
	}
	
	filtered, err = storage.GetComicsByFilter(filter)
	if err != nil {
		t.Errorf("GetComicsByFilter returned an error: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("Expected 1 comic matching Batman #001, got %d", len(filtered))
	}
}
