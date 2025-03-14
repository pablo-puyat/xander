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
}

func TestSQLiteStorage_Close(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()
	
	// Test closing the database
	err := storage.Close()
	assert.NoError(t, err)
}
