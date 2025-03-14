package storage

import (
	"time"
	"xander/internal/comicvine"
)

// Storage represents an interface for storing and retrieving comic metadata
type Storage interface {
	// StoreComic saves comic metadata to storage
	StoreComic(result *comicvine.Result) error
	
	// GetComics retrieves all stored comics
	GetComics() ([]*comicvine.Result, error)
	
	// GetComicByID retrieves a specific comic by its ID
	GetComicByID(id int) (*comicvine.Result, error)
	
	// GetComicsByFilter retrieves comics matching the provided filter criteria
	GetComicsByFilter(filter ComicFilter) ([]*comicvine.Result, error)
	
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

// NewFilter creates a new ComicFilter with default values
func NewFilter() ComicFilter {
	return ComicFilter{
		Limit:  100,
		Offset: 0,
	}
}