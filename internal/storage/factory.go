package storage

import (
	"fmt"
)

// StorageType represents the type of storage to use
type StorageType string

const (
	// SQLite storage type
	SQLite StorageType = "sqlite"
	// Memory storage type (for testing)
	Memory StorageType = "memory"
)

// GetStorage returns a storage instance of the requested type
func GetStorage(storageType StorageType, dbPath string) (Storage, error) {
	switch storageType {
	case SQLite:
		return NewSQLiteStorage(dbPath)
	case Memory:
		// Not implemented yet, but could be useful for testing
		return nil, fmt.Errorf("memory storage not implemented")
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}