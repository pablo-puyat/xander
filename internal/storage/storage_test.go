package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
