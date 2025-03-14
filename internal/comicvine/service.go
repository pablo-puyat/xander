package comicvine

import (
	"fmt"
	"log"
	"strings"
	"xander/internal/comic"
	"xander/internal/parse"
)

// ComicService represents a service for comic metadata operations
type ComicService struct {
	client *Client
	verbose bool
}

// NewComicService creates a new comic service
func NewComicService(apiKey string, verbose bool) *ComicService {
	return &ComicService{
		client: NewClient(apiKey, verbose),
		verbose: verbose,
	}
}

// Result represents the metadata result for a comic file
type Result struct {
	Filename    string
	Series      string
	Issue       string
	Year        string
	Publisher   string
	ComicVineID int
	Title       string
	CoverURL    string
	Description string
}

// GetMetadata retrieves metadata for a comic file
func (s *ComicService) GetMetadata(filename string) (*Result, error) {
	// Extract info from the filename
	series, issue, year, publisher, err := parse.ParseComicFilename(filename)
	if err != nil {
		if s.verbose {
			log.Printf("Failed to parse filename '%s': %v", filename, err)
		}
		return nil, fmt.Errorf("failed to parse filename: %w", err)
	}

	if s.verbose {
		log.Printf("Parsed '%s' as Series='%s', Issue='%s', Year='%s', Publisher='%s'", 
			filename, series, issue, year, publisher)
	}

	// Get issue from ComicVine
	comicInfo, err := s.client.GetIssue(series, issue)
	if err != nil {
		if s.verbose {
			log.Printf("Failed to get issue from ComicVine for '%s' #%s: %v", series, issue, err)
		}
		return nil, fmt.Errorf("failed to get issue from ComicVine: %w", err)
	}

	if s.verbose {
		log.Printf("Successfully retrieved metadata for '%s' #%s", series, issue)
	}

	// Create the result
	result := &Result{
		Filename:    filename,
		Series:      series,
		Issue:       issue,
		Year:        year,
		Publisher:   publisher,
		ComicVineID: comicInfo.ID,
		Title:       comicInfo.Name,
		CoverURL:    comicInfo.Image.OriginalURL,
		Description: comicInfo.Description,
	}

	return result, nil
}

// GetMetadataForFiles processes multiple files and returns their metadata
func (s *ComicService) GetMetadataForFiles(filenames []string) ([]*Result, error) {
	var results []*Result
	var errors []string

	for _, filename := range filenames {
		result, err := s.GetMetadata(filename)
		if err != nil {
			// Log error but continue with other files
			errorMsg := fmt.Sprintf("Error processing %s: %v", filename, err)
			fmt.Println(errorMsg)
			errors = append(errors, errorMsg)
			continue
		}

		results = append(results, result)
	}

	// Only return error if no results were found
	if len(results) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("failed to process any files: %s", strings.Join(errors, "; "))
	}

	return results, nil
}

// ToComic converts a ComicVine API Result to a domain Comic model
func (r *Result) ToComic() *comic.Comic {
	return &comic.Comic{
		Filename:    r.Filename,
		Series:      r.Series,
		Issue:       r.Issue,
		Year:        r.Year,
		Publisher:   r.Publisher,
		ComicVineID: r.ComicVineID,
		Title:       r.Title,
		CoverURL:    r.CoverURL,
		Description: r.Description,
	}
}

// FromComic creates a ComicVine Result from a domain Comic model
func FromComic(c *comic.Comic) *Result {
	return &Result{
		Filename:    c.Filename,
		Series:      c.Series,
		Issue:       c.Issue,
		Year:        c.Year,
		Publisher:   c.Publisher,
		ComicVineID: c.ComicVineID,
		Title:       c.Title,
		CoverURL:    c.CoverURL,
		Description: c.Description,
	}
}