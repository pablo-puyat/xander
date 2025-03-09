package comicvine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var baseURL = "https://comicvine.gamespot.com/api"

// Client handles requests to the ComicVine API
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new ComicVine API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// Image represents an image in the ComicVine API
type Image struct {
	OriginalURL string `json:"original_url"`
}

// Volume represents a comic volume in the ComicVine API
type Volume struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Issue represents a comic issue in the ComicVine API
type Issue struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	IssueNumber string `json:"issue_number"`
	Volume      Volume `json:"volume"`
	CoverDate   string `json:"cover_date"`
	Image       Image  `json:"image"`
	Description string `json:"description"`
}

// Response represents the API response from ComicVine
type Response struct {
	StatusCode int     `json:"status_code"`
	Error      string  `json:"error"`
	Results    []Issue `json:"results"`
}

// GetIssue searches for a comic issue
func (c *Client) GetIssue(series string, issueNumber string) (*Issue, error) {
	// Build the URL
	params := url.Values{}
	params.Add("api_key", c.apiKey)
	params.Add("format", "json")
	params.Add("filter", fmt.Sprintf("volume:%s,issue_number:%s", series, issueNumber))

	// Make the request
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/issues?%s", baseURL, params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.StatusCode != 1 {
		return nil, fmt.Errorf("API error: %s", result.Error)
	}

	if len(result.Results) == 0 {
		return nil, fmt.Errorf("no results found for %s #%s", series, issueNumber)
	}

	return &result.Results[0], nil
}
