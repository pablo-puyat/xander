package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFilter(t *testing.T) {
	filter := NewFilter()
	
	// Check default values
	assert.Equal(t, 100, filter.Limit)
	assert.Equal(t, 0, filter.Offset)
	assert.Equal(t, "", filter.Series)
	assert.Equal(t, "", filter.Issue)
	assert.Equal(t, "", filter.Year)
	assert.Equal(t, "", filter.Publisher)
	assert.Equal(t, "", filter.Filename)
	assert.True(t, filter.StartDate.IsZero())
	assert.True(t, filter.EndDate.IsZero())
}

func TestGetStorage(t *testing.T) {
	tests := []struct {
		name        string
		storageType StorageType
		dbPath      string
		expectErr   bool
		errContains string
	}{
		{
			name:        "SQLite storage with custom path",
			storageType: SQLite,
			dbPath:      ":memory:", // Use in-memory SQLite for testing
			expectErr:   false,
		},
		{
			name:        "Memory storage",
			storageType: Memory,
			dbPath:      "",
			expectErr:   false,
		},
		{
			name:        "Unknown storage type",
			storageType: StorageType("unknown"),
			dbPath:      "",
			expectErr:   true,
			errContains: "unknown storage type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := GetStorage(tt.storageType, tt.dbPath)
			
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
				
				// Cleanup
				if storage != nil {
					_ = storage.Close()
				}
			}
		})
	}
}