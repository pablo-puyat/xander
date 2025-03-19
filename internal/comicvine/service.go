package comicvine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type APIClient interface {
	Request(ctx context.Context, endpoint string, params map[string]string) ([]byte, int, error)
}

type ComicService struct {
	client  APIClient
	verbose bool
}

func NewService(apiKey string, verbose bool) *ComicService {
	return &ComicService{
		client:  NewClient(apiKey, verbose),
		verbose: verbose,
	}
}

type VolumeResult struct {
	ComicVineID int
	Name        string
	Publisher   string
	StartYear   string

	RawData map[string]interface{} // Complete raw response
}

type IssueResult struct {
	// Basic ComicVine data
	ComicVineID int
	Name        string
	IssueNumber string
	VolumeID    int
	Publisher   string // Publisher from API

	// Full raw data
	RawData map[string]interface{} // Complete raw response
}

// searchSeries is an internal method to search for series by name and optionally year
func (s *ComicService) searchSeries(ctx context.Context, name string, year string) ([]VolumeResult, error) {
	// Create parameters map for the API request
	params := map[string]string{
		"resources": "volume", // Search for comic volumes/series
	}

	// Add filter for series name and optionally year
	filterValue := "name:" + name
	if year != "" {
		filterValue += ",start_year:" + year
	}
	params["filter"] = filterValue

	// Make the API request
	response, code, err := s.client.Request(ctx, "search", params)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	// Check status code
	if code != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", code)
	}

	// Parse the response
	var apiResponse struct {
		Results []struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			StartYear string `json:"start_year"`
			Publisher struct {
				Name string `json:"name"`
			} `json:"publisher"`
			Image struct {
				OriginalURL string `json:"original_url"`
			} `json:"image"`
		} `json:"results"`
	}

	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Convert API results to our result type
	results := make([]VolumeResult, 0, len(apiResponse.Results))
	for _, item := range apiResponse.Results {
		results = append(results, VolumeResult{
			ComicVineID: item.ID,
			Name:        item.Name,
			StartYear:   item.StartYear,
			Publisher:   item.Publisher.Name,
		})
	}

	return results, nil
}

func (s *ComicService) getIssue(ctx context.Context, volumeId int, issue string) ([]IssueResult, error) {
	params := map[string]string{
		"resources": "issue",
		"filter":    "volume:" + strconv.Itoa(volumeId) + ",issue_number:" + issue,
	}
	response, code, err := s.client.Request(ctx, "search", params)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	// Check status code
	if code != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", code)
	}

	// Parse the response
	var apiResponse struct {
		Results []struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Issue     string `json:"issue"`
			Publisher struct {
				Name string `json:"name"`
			} `json:"publisher"`
			Image struct {
				OriginalURL string `json:"original_url"`
			} `json:"image"`
		} `json:"results"`
	}

	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Convert API results to our result type
	results := make([]IssueResult, 0, len(apiResponse.Results))
	for _, item := range apiResponse.Results {
		results = append(results, IssueResult{
			ComicVineID: item.ID,
			Name:        item.Name,
			Publisher:   item.Publisher.Name,
			IssueNumber: item.Issue,
		})
	}

	return results, nil
}

// SearchSeries searches for comic series by name and optional year
// It returns a list of matching series with their metadata
func (s *ComicService) SearchSeries(ctx context.Context, name string, year string) ([]VolumeResult, error) {
	if s.verbose {
		log.Printf("Searching for series: %s, year: %s", name, year)
	}

	return s.searchSeries(ctx, name, year)
}
