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
	
	// Create test data
	testComic := createTestComic()
	
	// Test with SQLite
	err = sqliteStorage.StoreComic(testComic)
	assert.NoError(t, err)
	
	// Cleanup
	err = sqliteStorage.Close()
	assert.NoError(t, err)
}
