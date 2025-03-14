package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStorageWithIntegration(t *testing.T) {
	// This test is more of an integration test that checks if
	// both implementations satisfy the Storage interface correctly
	
	// Create an SQLite storage
	sqliteStorage, err := GetStorage(SQLite, ":memory:")
	assert.NoError(t, err)
	assert.NotNil(t, sqliteStorage)
	
	// Create a Memory storage
	memoryStorage, err := GetStorage(Memory, "")
	assert.NoError(t, err)
	assert.NotNil(t, memoryStorage)
	
	// Create test data
	testComic := createTestComic()
	
	// Test with SQLite
	err = sqliteStorage.StoreComic(testComic)
	assert.NoError(t, err)
	
	sqliteResult, err := sqliteStorage.GetComicByID(testComic.ComicVineID)
	assert.NoError(t, err)
	assert.NotNil(t, sqliteResult)
	assert.Equal(t, testComic.Series, sqliteResult.Series)
	
	// Test with Memory
	err = memoryStorage.StoreComic(testComic)
	assert.NoError(t, err)
	
	memoryResult, err := memoryStorage.GetComicByID(testComic.ComicVineID)
	assert.NoError(t, err)
	assert.NotNil(t, memoryResult)
	assert.Equal(t, testComic.Series, memoryResult.Series)
	
	// Cleanup
	err = sqliteStorage.Close()
	assert.NoError(t, err)
	
	err = memoryStorage.Close()
	assert.NoError(t, err)
}