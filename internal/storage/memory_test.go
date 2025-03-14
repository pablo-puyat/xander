package storage

import (
	"testing"
	"xander/internal/comicvine"

	"github.com/stretchr/testify/assert"
)

// TestMemoryStorage tests the memory storage implementation
func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()
	
	// Create test comics
	comic1 := createTestComic()
	comic2 := &comicvine.Result{
		Filename:    "test-comic-2.cbz",
		Series:      "Another Series",
		Issue:       "2",
		Year:        "2024",
		Publisher:   "Another Publisher",
		ComicVineID: 654321,
		Title:       "Another Comic",
	}
	
	// Test StoreComic
	err := storage.StoreComic(comic1)
	assert.NoError(t, err)
	err = storage.StoreComic(comic2)
	assert.NoError(t, err)
	
	// Test GetComics
	comics, err := storage.GetComics()
	assert.NoError(t, err)
	assert.Len(t, comics, 2)
	
	// Test GetComicByID
	result, err := storage.GetComicByID(comic1.ComicVineID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, comic1.ComicVineID, result.ComicVineID)
	
	// Test GetComicsByFilter
	filter := ComicFilter{
		Series: "Another",
		Limit:  100,
	}
	
	filteredResults, err := storage.GetComicsByFilter(filter)
	assert.NoError(t, err)
	assert.Len(t, filteredResults, 1) // Should find one match with partial match
	
	// Test Close
	err = storage.Close()
	assert.NoError(t, err)
}