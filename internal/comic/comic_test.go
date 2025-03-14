package comic

import (
	"testing"
)

func TestComic(t *testing.T) {
	c := &Comic{
		Filename:    "X-Men.001.cbz",
		Series:      "X-Men",
		Issue:       "001",
		Year:        "2021",
		Publisher:   "Marvel",
		ComicVineID: 123456,
		Title:       "Test Title",
		CoverURL:    "http://example.com/cover.jpg",
		Description: "Test description",
	}

	if c.Filename != "X-Men.001.cbz" {
		t.Errorf("Expected Filename to be 'X-Men.001.cbz', got '%s'", c.Filename)
	}
	if c.Series != "X-Men" {
		t.Errorf("Expected Series to be 'X-Men', got '%s'", c.Series)
	}
	if c.Issue != "001" {
		t.Errorf("Expected Issue to be '001', got '%s'", c.Issue)
	}
	if c.Year != "2021" {
		t.Errorf("Expected Year to be '2021', got '%s'", c.Year)
	}
	if c.Publisher != "Marvel" {
		t.Errorf("Expected Publisher to be 'Marvel', got '%s'", c.Publisher)
	}
	if c.ComicVineID != 123456 {
		t.Errorf("Expected ComicVineID to be 123456, got %d", c.ComicVineID)
	}
	if c.Title != "Test Title" {
		t.Errorf("Expected Title to be 'Test Title', got '%s'", c.Title)
	}
	if c.CoverURL != "http://example.com/cover.jpg" {
		t.Errorf("Expected CoverURL to be 'http://example.com/cover.jpg', got '%s'", c.CoverURL)
	}
	if c.Description != "Test description" {
		t.Errorf("Expected Description to be 'Test description', got '%s'", c.Description)
	}
}

func TestFilter(t *testing.T) {
	f := Filter{
		Series:    "X-Men",
		Issue:     "001",
		Year:      "2021",
		Publisher: "Marvel",
		Filename:  "X-Men.001.cbz",
	}

	if f.Series != "X-Men" {
		t.Errorf("Expected Series to be 'X-Men', got '%s'", f.Series)
	}
	if f.Issue != "001" {
		t.Errorf("Expected Issue to be '001', got '%s'", f.Issue)
	}
	if f.Year != "2021" {
		t.Errorf("Expected Year to be '2021', got '%s'", f.Year)
	}
	if f.Publisher != "Marvel" {
		t.Errorf("Expected Publisher to be 'Marvel', got '%s'", f.Publisher)
	}
	if f.Filename != "X-Men.001.cbz" {
		t.Errorf("Expected Filename to be 'X-Men.001.cbz', got '%s'", f.Filename)
	}
}