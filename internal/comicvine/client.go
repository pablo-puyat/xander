package comicvine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var baseURL = "https://comicvine.gamespot.com/api"

// Client handles requests to the ComicVine API
type Client struct {
	apiKey     string
	httpClient *http.Client
	verbose    bool
}

// NewClient creates a new ComicVine API client
func NewClient(apiKey string, verbose bool) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		verbose: verbose,
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
	// Normalize issue number (remove leading zeros)
	normalizedIssueNumber := issueNumber
	if len(issueNumber) > 1 && issueNumber[0] == '0' {
		// Strip leading zeros
		for len(normalizedIssueNumber) > 1 && normalizedIssueNumber[0] == '0' {
			normalizedIssueNumber = normalizedIssueNumber[1:]
		}
	}

	if c.verbose {
		log.Printf("Searching for series: '%s', issue: '%s' (normalized: '%s')", 
			series, issueNumber, normalizedIssueNumber)
	}

	// Build the URL
	params := url.Values{}
	params.Add("api_key", c.apiKey)
	params.Add("format", "json")
	params.Add("limit", "10")          // Increase result count
	params.Add("field_list", "id,name,issue_number,volume,cover_date,image,description") // Only get fields we need
	
	// Use query parameter (more flexible than filter)
	query := fmt.Sprintf("%s %s", series, normalizedIssueNumber)
	params.Add("query", query)
	
	// Add sort to get most relevant results first
	params.Add("sort", "name:asc")
	
	requestURL := fmt.Sprintf("%s/issues?%s", baseURL, params.Encode())
	
	// Log the request URL (with API key masked)
	if c.verbose {
		maskedParams := url.Values{}
		for k, v := range params {
			if k == "api_key" {
				maskedParams.Add(k, "********")
			} else {
				maskedParams[k] = v
			}
		}
		maskedURL := fmt.Sprintf("%s/issues?%s", baseURL, maskedParams.Encode())
		log.Printf("ComicVine API Request: GET %s", maskedURL)
	}

	// Make the request
	resp, err := c.httpClient.Get(requestURL)
	if err != nil {
		if c.verbose {
			log.Printf("ComicVine API Error: %v", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if c.verbose {
			log.Printf("ComicVine API Error reading response: %v", err)
		}
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log the response
	if c.verbose {
		log.Printf("ComicVine API Response Status: %s", resp.Status)
		if len(body) > 1000 {
			log.Printf("ComicVine API Response (truncated): %s...", body[:1000])
		} else {
			log.Printf("ComicVine API Response: %s", body)
		}
	}

	// Parse the response
	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		if c.verbose {
			log.Printf("ComicVine API Error parsing JSON: %v", err)
		}
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.StatusCode != 1 {
		if c.verbose {
			log.Printf("ComicVine API Error: Status code %d, Error: %s", result.StatusCode, result.Error)
		}
		return nil, fmt.Errorf("API error: %s", result.Error)
	}

	if len(result.Results) == 0 {
		if c.verbose {
			log.Printf("ComicVine API: No results found for series '%s', issue '%s'", series, issueNumber)
		}
		return nil, fmt.Errorf("no results found for %s #%s", series, issueNumber)
	}

	// Log all results we found
	if c.verbose {
		log.Printf("ComicVine API: Found %d results for query '%s %s'", len(result.Results), series, issueNumber)
		for i, issue := range result.Results {
			log.Printf("  Result %d: ID=%d, Name='%s', Volume='%s', Issue='%s'", 
				i+1, issue.ID, issue.Name, issue.Volume.Name, issue.IssueNumber)
		}
	}

	// Try to find the best match
	var bestMatch *Issue
	bestScore := -1

	for i := range result.Results {
		issue := &result.Results[i]
		score := 0
		
		// If the volume name contains our series name, that's good
		if strings.Contains(strings.ToLower(issue.Volume.Name), strings.ToLower(series)) {
			score += 5
		}
		
		// If the issue number matches exactly, that's very good
		if issue.IssueNumber == issueNumber || issue.IssueNumber == normalizedIssueNumber {
			score += 10
		}
		
		if score > bestScore {
			bestScore = score
			bestMatch = issue
		}
	}
	
	// If we couldn't find a good match, just use the first result
	if bestMatch == nil {
		bestMatch = &result.Results[0]
	}

	if c.verbose {
		log.Printf("ComicVine API: Best match for '%s #%s': ID=%d, Name='%s', Volume='%s', Issue='%s'", 
			series, issueNumber, bestMatch.ID, bestMatch.Name, bestMatch.Volume.Name, bestMatch.IssueNumber)
	}

	return bestMatch, nil
}