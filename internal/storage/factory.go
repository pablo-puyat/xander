package storage

import (
	"fmt"
)

// StorageType represents the type of storage to use
type StorageType string

const (
	// SQLite storage type
	SQLite StorageType = "sqlite"
)

// GetStorage returns a storage instance of the requested type
func GetStorage(storageType StorageType, dbPath string) (Storage, error) {
	switch storageType {
	case SQLite:
		return NewSQLiteStorage(dbPath)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}
