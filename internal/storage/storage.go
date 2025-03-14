package storage

import (
	"time"
	"xander/internal/comicvine"
)

// Storage represents an interface for storing and retrieving comic metadata
type Storage interface {
	// StoreComic saves comic metadata to storage
	StoreComic(result *comicvine.Result) error
	
	// Close closes the storage connection
	Close() error
}

// ComicFilter defines criteria for filtering comics
type ComicFilter struct {
	Series    string
	Issue     string
	Year      string
	Publisher string
	Filename  string
	StartDate time.Time
	EndDate   time.Time
	Limit     int
	Offset    int
}

