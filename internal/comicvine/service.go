package comicvine

import (
	"fmt"
	"xander/internal/parse"
)

// ComicService represents a service for comic metadata operations
type ComicService struct {
	client *Client
}

// NewComicService creates a new comic service
func NewComicService(apiKey string) *ComicService {
	return &ComicService{
		client: NewClient(apiKey),
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
		return nil, fmt.Errorf("failed to parse filename: %w", err)
	}

	// Get issue from ComicVine
	comicInfo, err := s.client.GetIssue(series, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue from ComicVine: %w", err)
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