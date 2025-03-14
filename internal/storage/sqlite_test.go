package storage

import (
	"os"
	"path/filepath"
	"testing"
	"xander/internal/comicvine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestComic() *comicvine.Result {
	return &comicvine.Result{
		Filename:    "test-comic.cbz",
		Series:      "Test Series",
		Issue:       "1",
		Year:        "2023",
		Publisher:   "Test Publisher",
		ComicVineID: 123456,
		Title:       "Test Comic",
		CoverURL:    "http://example.com/cover.jpg",
		Description: "Test description",
	}
}

func setupTestDB(t *testing.T) (*SQLiteStorage, func()) {
	// Create a temporary file for the test database
	tempDir, err := os.MkdirTemp("", "xander-test")
	require.NoError(t, err)
	
	dbPath := filepath.Join(tempDir, "test.db")
	
	// Create the SQLite storage
	storage, err := NewSQLiteStorage(dbPath)
	require.NoError(t, err)
	require.NotNil(t, storage)
	
	// Return the storage and a cleanup function
	return storage, func() {
		_ = storage.Close()
		_ = os.RemoveAll(tempDir)
	}
}

func TestNewSQLiteStorage(t *testing.T) {
	tests := []struct {
		name      string
		dbPath    string
		expectErr bool
	}{
		{
			name:      "Custom path",
			dbPath:    ":memory:", // Use in-memory SQLite for testing
			expectErr: false,
		},
		{
			name:      "Default path (uses home directory)",
			dbPath:    "",
			expectErr: false,
		},
		{
			name:      "Invalid path",
			dbPath:    "/path/that/cannot/exist/for/testing", // This should fail on permissions
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var storage *SQLiteStorage
			var err error
			
			if tt.dbPath == "" {
				// Skip the default path test if we can't determine home directory
				_, err = os.UserHomeDir()
				if err != nil {
					t.Skip("Skipping test that requires home directory")
				}
			}
			
			storage, err = NewSQLiteStorage(tt.dbPath)
			
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
				
				// Check that tables were created
				var count int
				err = storage.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='comics'").Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 1, count, "Comics table should exist")
				
				// Cleanup
				if storage != nil {
					_ = storage.Close()
				}
				
				// Cleanup default path if it was created
				if tt.dbPath == "" {
					userHome, _ := os.UserHomeDir()
					dbPath := filepath.Join(userHome, ".config", "xander", "xander.db")
					_ = os.Remove(dbPath)
				}
			}
		})
	}
}

func TestSQLiteStorage_StoreComic(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()
	
	// Create a test comic
	comic := createTestComic()
	
	// Store the comic
	err := storage.StoreComic(comic)
	assert.NoError(t, err)
	
	// Verify the comic was stored
	var count int
	err = storage.db.QueryRow("SELECT COUNT(*) FROM comics WHERE comicvine_id = ?", comic.ComicVineID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	
	// Update the comic
	comic.Title = "Updated Title"
	err = storage.StoreComic(comic)
	assert.NoError(t, err)
	
	// Verify the comic was updated
	var title string
	err = storage.db.QueryRow("SELECT title FROM comics WHERE comicvine_id = ?", comic.ComicVineID).Scan(&title)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", title)
}

func TestSQLiteStorage_GetComicByID(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()
	
	// Create and store a test comic
	comic := createTestComic()
	err := storage.StoreComic(comic)
	assert.NoError(t, err)
	
	// Get the comic by ID
	result, err := storage.GetComicByID(comic.ComicVineID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	// Verify the comic data
	assert.Equal(t, comic.ComicVineID, result.ComicVineID)
	assert.Equal(t, comic.Series, result.Series)
	assert.Equal(t, comic.Issue, result.Issue)
	assert.Equal(t, comic.Year, result.Year)
	assert.Equal(t, comic.Publisher, result.Publisher)
	assert.Equal(t, comic.Title, result.Title)
	assert.Equal(t, comic.CoverURL, result.CoverURL)
	assert.Equal(t, comic.Description, result.Description)
	assert.Equal(t, comic.Filename, result.Filename)
	
	// Test non-existent comic
	result, err = storage.GetComicByID(999999)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestSQLiteStorage_GetComics(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()
	
	// Create and store multiple test comics
	comic1 := createTestComic()
	comic2 := &comicvine.Result{
		Filename:    "test-comic-2.cbz",
		Series:      "Another Series",
		Issue:       "2",
		Year:        "2024",
		Publisher:   "Another Publisher",
		ComicVineID: 654321,
		Title:       "Another Comic",
		CoverURL:    "http://example.com/cover2.jpg",
		Description: "Another description",
	}
	
	err := storage.StoreComic(comic1)
	assert.NoError(t, err)
	err = storage.StoreComic(comic2)
	assert.NoError(t, err)
	
	// Get all comics
	results, err := storage.GetComics()
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestSQLiteStorage_GetComicsByFilter(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()
	
	// Create and store multiple test comics
	comics := []*comicvine.Result{
		{
			Filename:    "batman-1.cbz",
			Series:      "Batman",
			Issue:       "1",
			Year:        "2020",
			Publisher:   "DC Comics",
			ComicVineID: 111111,
			Title:       "Batman #1",
		},
		{
			Filename:    "batman-2.cbz",
			Series:      "Batman",
			Issue:       "2",
			Year:        "2020",
			Publisher:   "DC Comics",
			ComicVineID: 222222,
			Title:       "Batman #2",
		},
		{
			Filename:    "superman-1.cbz",
			Series:      "Superman",
			Issue:       "1",
			Year:        "2021",
			Publisher:   "DC Comics",
			ComicVineID: 333333,
			Title:       "Superman #1",
		},
		{
			Filename:    "amazing-spider-man-1.cbz",
			Series:      "Amazing Spider-Man",
			Issue:       "1",
			Year:        "2022",
			Publisher:   "Marvel",
			ComicVineID: 444444,
			Title:       "Amazing Spider-Man #1",
		},
	}
	
	for _, comic := range comics {
		err := storage.StoreComic(comic)
		assert.NoError(t, err)
	}
	
	// Test filters
	tests := []struct {
		name       string
		filter     ComicFilter
		expectLen  int
		expectIDs  []int
	}{
		{
			name: "Filter by series",
			filter: ComicFilter{
				Series: "Batman",
				Limit:  100,
			},
			expectLen: 2,
			expectIDs: []int{111111, 222222},
		},
		{
			name: "Filter by issue",
			filter: ComicFilter{
				Issue: "1",
				Limit: 100,
			},
			expectLen: 3,
			expectIDs: []int{111111, 333333, 444444},
		},
		{
			name: "Filter by year",
			filter: ComicFilter{
				Year:  "2020",
				Limit: 100,
			},
			expectLen: 2,
			expectIDs: []int{111111, 222222},
		},
		{
			name: "Filter by publisher",
			filter: ComicFilter{
				Publisher: "Marvel",
				Limit:     100,
			},
			expectLen: 1,
			expectIDs: []int{444444},
		},
		{
			name: "Filter by filename",
			filter: ComicFilter{
				Filename: "spider",
				Limit:    100,
			},
			expectLen: 1,
			expectIDs: []int{444444},
		},
		{
			name: "Combined filters",
			filter: ComicFilter{
				Series:    "Batman",
				Issue:     "1",
				Publisher: "DC",
				Limit:     100,
			},
			expectLen: 1,
			expectIDs: []int{111111},
		},
		{
			name: "Limit and offset",
			filter: ComicFilter{
				Limit:  2,
				Offset: 1,
			},
			expectLen: 2,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := storage.GetComicsByFilter(tt.filter)
			assert.NoError(t, err)
			assert.Len(t, results, tt.expectLen)
			
			if len(tt.expectIDs) > 0 {
				ids := make([]int, len(results))
				for i, result := range results {
					ids[i] = result.ComicVineID
				}
				
				// Check that all expected IDs are present
				for _, id := range tt.expectIDs {
					found := false
					for _, resultID := range ids {
						if resultID == id {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected ID %d not found in results", id)
				}
			}
		})
	}
	
	// Test without date filters - should return all comics
	basicFilter := ComicFilter{
		Limit:  100,
		Offset: 0,
	}
	
	results, err := storage.GetComicsByFilter(basicFilter)
	assert.NoError(t, err)
	assert.Len(t, results, 4) // All comics should be returned
}

func TestSQLiteStorage_Close(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()
	
	// Test closing the database
	err := storage.Close()
	assert.NoError(t, err)
	
	// Verify the database is closed by attempting an operation that should fail
	_, err = storage.GetComics()
	assert.Error(t, err)
}