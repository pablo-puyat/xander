package storage

import (
	"strings"
	"xander/internal/comicvine"
)

// MemoryStorage implements the Storage interface using in-memory maps
// This is primarily for testing purposes
type MemoryStorage struct {
	comics map[int]*comicvine.Result
}

// NewMemoryStorage creates a new memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		comics: make(map[int]*comicvine.Result),
	}
}

// StoreComic saves comic metadata to storage
func (m *MemoryStorage) StoreComic(result *comicvine.Result) error {
	m.comics[result.ComicVineID] = result
	return nil
}

// GetComics retrieves all stored comics
func (m *MemoryStorage) GetComics() ([]*comicvine.Result, error) {
	filter := NewFilter()
	return m.GetComicsByFilter(filter)
}

// GetComicByID retrieves a specific comic by its ID
func (m *MemoryStorage) GetComicByID(id int) (*comicvine.Result, error) {
	if comic, exists := m.comics[id]; exists {
		return comic, nil
	}
	return nil, nil
}

// GetComicsByFilter retrieves comics matching the provided filter criteria
func (m *MemoryStorage) GetComicsByFilter(filter ComicFilter) ([]*comicvine.Result, error) {
	var results []*comicvine.Result
	
	// Apply filters
	for _, comic := range m.comics {
		if m.matchesFilter(comic, filter) {
			results = append(results, comic)
		}
	}
	
	// Apply limit and offset
	if filter.Offset < len(results) {
		end := filter.Offset + filter.Limit
		if end > len(results) {
			end = len(results)
		}
		results = results[filter.Offset:end]
	} else {
		results = []*comicvine.Result{}
	}
	
	return results, nil
}

// matchesFilter checks if a comic matches the filter criteria
func (m *MemoryStorage) matchesFilter(comic *comicvine.Result, filter ComicFilter) bool {
	// Series filter
	if filter.Series != "" && !strings.Contains(strings.ToLower(comic.Series), strings.ToLower(filter.Series)) {
		return false
	}
	
	// Issue filter
	if filter.Issue != "" && comic.Issue != filter.Issue {
		return false
	}
	
	// Year filter
	if filter.Year != "" && comic.Year != filter.Year {
		return false
	}
	
	// Publisher filter
	if filter.Publisher != "" && !strings.Contains(strings.ToLower(comic.Publisher), strings.ToLower(filter.Publisher)) {
		return false
	}
	
	// Filename filter
	if filter.Filename != "" && !strings.Contains(strings.ToLower(comic.Filename), strings.ToLower(filter.Filename)) {
		return false
	}
	
	return true
}

// Close closes the storage connection
func (m *MemoryStorage) Close() error {
	// Clear the map to free memory
	m.comics = make(map[int]*comicvine.Result)
	return nil
}