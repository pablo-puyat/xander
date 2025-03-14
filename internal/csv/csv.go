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

// For backward compatibility - maps to domain Filter
type CSVFilter = comic.Filter

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
	var err error
	
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
	
	// Load existing comics if file exists and is not empty
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	
	if fileInfo.Size() > 0 {
		// Load existing comics
		comics, err := ReadCSV(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read existing comics: %w", err)
		}
		storage.comics = comics
		
		// Reset file position to beginning
		_, err = file.Seek(0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to reset file position: %w", err)
		}
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

// GetComics retrieves all comics from the CSV storage
func (s *CSVStorage) GetComics() ([]*comic.Comic, error) {
	return s.comics, nil
}

// GetComicByID retrieves a comic by its ComicVine ID
func (s *CSVStorage) GetComicByID(id int) (*comic.Comic, error) {
	for _, c := range s.comics {
		if c.ComicVineID == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("comic with ID %d not found", id)
}

// GetComicsByFilter retrieves comics matching the filter criteria
func (s *CSVStorage) GetComicsByFilter(filter comic.Filter) ([]*comic.Comic, error) {
	var results []*comic.Comic
	
	for _, c := range s.comics {
		// Check if comic matches all the provided filter criteria
		if (filter.Series == "" || c.Series == filter.Series) &&
		   (filter.Issue == "" || c.Issue == filter.Issue) &&
		   (filter.Year == "" || c.Year == filter.Year) &&
		   (filter.Publisher == "" || c.Publisher == filter.Publisher) &&
		   (filter.Filename == "" || c.Filename == filter.Filename) {
			results = append(results, c)
		}
	}
	
	return results, nil
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
    headers := []string{
        "Filename", "Series", "Issue", "Year", "Publisher", 
        "ComicVineID", "Title", "CoverURL", "Description",
    }
    
    // Create CSV writer
    var buf strings.Builder
    writer := csv.NewWriter(&buf)
    
    // Write headers
    if err := writer.Write(headers); err != nil {
        return "", fmt.Errorf("error writing CSV headers: %w", err)
    }
    
    // Write data rows
    for _, c := range comics {
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

// ReadCSV reads comics from a CSV reader
func ReadCSV(r io.Reader) ([]*comic.Comic, error) {
    reader := csv.NewReader(r)
    
    // Read all records
    records, err := reader.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("error reading CSV data: %w", err)
    }
    
    if len(records) < 2 { // Need at least headers and one data row
        return []*comic.Comic{}, nil
    }
    
    // Skip header row (first row)
    dataRows := records[1:]
    
    // Parse each row into a Comic
    comics := make([]*comic.Comic, 0, len(dataRows))
    for _, row := range dataRows {
        if len(row) < 9 { // Ensure we have all fields
            continue // Skip malformed rows
        }
        
        comicVineID, err := strconv.Atoi(row[5])
        if err != nil {
            comicVineID = 0 // Default value if conversion fails
        }
        
        comic := &comic.Comic{
            Filename:    row[0],
            Series:      row[1],
            Issue:       row[2],
            Year:        row[3],
            Publisher:   row[4],
            ComicVineID: comicVineID,
            Title:       row[6],
            CoverURL:    row[7],
            Description: row[8],
        }
        
        comics = append(comics, comic)
    }
    
    return comics, nil
}
