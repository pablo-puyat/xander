package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"xander/internal/comic"
)

// CSVStorage implements comic.Storage interface for CSV-based persistence
type CSVStorage struct {
	filepath string
	file     *os.File
	comics   []*comic.Comic
}

// Ensure CSVStorage implements comic.Storage interface
var _ comic.Storage = (*CSVStorage)(nil)

// NewCSVStorage creates a new CSV storage at the specified filepath
func NewCSVStorage(filepath string) (*CSVStorage, error) {
	// Check if file exists
	var file *os.File
	
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// Create new file if it doesn't exist
		file, err = os.Create(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create CSV file: %w", err)
		}
	} else {
		// Open existing file if it exists
		file, err = os.OpenFile(filepath, os.O_RDWR, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open CSV file: %w", err)
		}
	}
	
	storage := &CSVStorage{
		filepath: filepath,
		file:     file,
		comics:   []*comic.Comic{},
	}
	
	return storage, nil
}

// StoreComic saves a comic to the CSV storage
func (s *CSVStorage) StoreComic(comic *comic.Comic) error {
	// Add to in-memory collection
	s.comics = append(s.comics, comic)
	
	// Write to file
	err := WriteCSV(s.file, s.comics)
	if err != nil {
		return fmt.Errorf("failed to write comic to CSV: %w", err)
	}
	
	return nil
}

// Close closes the CSV storage and writes any pending changes
func (s *CSVStorage) Close() error {
	return s.file.Close()
}

// ComicToCSV converts domain comic objects to a CSV string
func ComicToCSV(comics []*comic.Comic) (string, error) {
    if len(comics) == 0 {
        return "", nil
    }
    
    // Define CSV headers based on Comic struct
    headers := []string{"Filename", "Series", "Issue", "Year", "Publisher", 
        "ComicVineID", "Title", "CoverURL", "Description",
        "StoreDate", "CoverDate", "DateAdded", "DateLastUpdated",
        "Characters", "Teams", "People"}
    
    // Create CSV writer
    var buf strings.Builder
    writer := csv.NewWriter(&buf)
    
    // Write headers
    if err := writer.Write(headers); err != nil {
        return "", fmt.Errorf("error writing CSV headers: %w", err)
    }
    
    // Write data rows
    for _, c := range comics {
        // Basic data
        row := []string{
            c.Filename,
            c.Series,
            c.Issue,
            c.Year,
            c.Publisher,
            strconv.Itoa(c.ComicVineID),
            c.Title,
            c.CoverURL,
            c.Description,
        }
        
        // Extended data
        // Add additional fields, handling nil values
        row = append(row, c.StoreDate)
        row = append(row, c.CoverDate)
        row = append(row, c.DateAdded)
        row = append(row, c.DateLastUpdated)
        
        // Convert array fields to comma-separated strings
        characterNames := extractNames(c.Characters)
        teamNames := extractNames(c.Teams)
        peopleNames := extractNames(c.People)
        
        row = append(row, characterNames)
        row = append(row, teamNames)
        row = append(row, peopleNames)
        
        if err := writer.Write(row); err != nil {
            return "", fmt.Errorf("error writing comic to CSV: %w", err)
        }
    }
    
    writer.Flush()
    if err := writer.Error(); err != nil {
        return "", fmt.Errorf("error flushing CSV data: %w", err)
    }
    
    return buf.String(), nil
}

// extractNames converts an array of entity maps to a comma-separated string of names
func extractNames(entities []map[string]interface{}) string {
	if entities == nil || len(entities) == 0 {
		return ""
	}
	
	var names []string
	for _, entity := range entities {
		if name, ok := entity["name"].(string); ok {
			names = append(names, name)
		}
	}
	
	return strings.Join(names, ", ")
}

// WriteCSV writes comics to a CSV writer
func WriteCSV(w io.Writer, comics []*comic.Comic) error {
    // Reset to beginning of file if it's a file
    if f, ok := w.(*os.File); ok {
        if _, err := f.Seek(0, 0); err != nil {
            return fmt.Errorf("failed to reset file position: %w", err)
        }
        // Truncate the file to clear previous content
        if err := f.Truncate(0); err != nil {
            return fmt.Errorf("failed to truncate file: %w", err)
        }
    }
    
    csvData, err := ComicToCSV(comics)
    if err != nil {
        return err
    }
    
    _, err = w.Write([]byte(csvData))
    return err
}

