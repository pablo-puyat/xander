package comicvine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var baseURL = "https://comicvine.gamespot.com/api"

// CacheEntry represents a cached API response
type CacheEntry struct {
	Results   []Issue
	Timestamp time.Time
}

// Client handles requests to the ComicVine API
type Client struct {
	apiKey               string
	httpClient           *http.Client
	verbose              bool
	lastRequestTime      time.Time
	requestCount         int
	requestCountResetTime time.Time
	cache                map[string]CacheEntry // Simple in-memory cache
}

// NewClient creates a new ComicVine API client
func NewClient(apiKey string, verbose bool) *Client {
	// Initialize with rate limiting support
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		verbose: verbose,
		lastRequestTime: time.Time{}, // Zero time
		requestCount: 0,
		requestCountResetTime: time.Now().Add(time.Hour), // Reset after 1 hour
		cache: make(map[string]CacheEntry), // Initialize cache
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

// checkAndRespectRateLimit enforces ComicVine's rate limits:
// - 200 requests per hour per resource
// - Adds delay between requests to prevent velocity detection
func (c *Client) checkAndRespectRateLimit() error {
	now := time.Now()
	
	// Check if we need to reset the counter (hourly)
	if now.After(c.requestCountResetTime) {
		if c.verbose {
			log.Printf("ComicVine API: Resetting rate limit counter (hourly reset)")
		}
		c.requestCount = 0
		c.requestCountResetTime = now.Add(time.Hour)
	}
	
	// Check if we've exceeded hourly limit (200 requests per hour)
	if c.requestCount >= 200 {
		resetTime := c.requestCountResetTime
		waitDuration := resetTime.Sub(now)
		
		if c.verbose {
			log.Printf("ComicVine API: Rate limit reached (200/hour). Need to wait %v until %v", 
				waitDuration.Round(time.Second), resetTime.Format(time.RFC3339))
		}
		
		// Return rate limit error
		return fmt.Errorf("ComicVine API rate limit exceeded (200/hour). Try again after %v", 
			resetTime.Format(time.RFC3339))
	}
	
	// Add delay between requests to prevent "velocity detection" blocks
	// We'll use a 1 second minimum delay between requests
	if !c.lastRequestTime.IsZero() {
		sinceLastRequest := now.Sub(c.lastRequestTime)
		requestDelay := 1 * time.Second // Minimum delay
		
		if sinceLastRequest < requestDelay {
			sleepDuration := requestDelay - sinceLastRequest
			if c.verbose {
				log.Printf("ComicVine API: Adding delay of %v between requests to prevent velocity detection", 
					sleepDuration.Round(time.Millisecond))
			}
			time.Sleep(sleepDuration)
		}
	}
	
	// Update request tracking
	c.lastRequestTime = time.Now()
	c.requestCount++
	
	if c.verbose {
		log.Printf("ComicVine API: Request %d of 200 for this hour (resets at %v)", 
			c.requestCount, c.requestCountResetTime.Format(time.RFC3339))
	}
	
	return nil
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
	
	// Create a cache key based on the series and issue number
	cacheKey := fmt.Sprintf("%s:%s", strings.ToLower(series), normalizedIssueNumber)
	
	// Check cache before making a request (cache entries valid for 24 hours)
	if entry, found := c.cache[cacheKey]; found {
		cacheTTL := 24 * time.Hour
		cacheAge := time.Since(entry.Timestamp)
		
		if cacheAge < cacheTTL {
			if c.verbose {
				log.Printf("ComicVine API: Using cached response for '%s #%s' (age: %v, expires in: %v)",
					series, issueNumber, cacheAge.Round(time.Second), (cacheTTL - cacheAge).Round(time.Second))
			}
			
			// Use cached results
			if len(entry.Results) == 0 {
				return nil, fmt.Errorf("no results found for %s #%s (cached response)", series, issueNumber)
			}
			
			// Find best match using same logic as below
			var bestMatch *Issue
			bestScore := -1
			
			for i := range entry.Results {
				issue := &entry.Results[i]
				score := 0
				
				if strings.Contains(strings.ToLower(issue.Volume.Name), strings.ToLower(series)) {
					score += 5
				}
				
				if issue.IssueNumber == issueNumber || issue.IssueNumber == normalizedIssueNumber {
					score += 10
				}
				
				if score > bestScore {
					bestScore = score
					bestMatch = issue
				}
			}
			
			// If no good match found, use first result
			if bestMatch == nil && len(entry.Results) > 0 {
				bestMatch = &entry.Results[0]
			}
			
			if bestMatch != nil {
				if c.verbose {
					log.Printf("ComicVine API: Using cached best match for '%s #%s': ID=%d, Name='%s'",
						series, issueNumber, bestMatch.ID, bestMatch.Name)
				}
				return bestMatch, nil
			}
		} else if c.verbose {
			log.Printf("ComicVine API: Cache entry expired for '%s #%s' (age: %v)",
				series, issueNumber, cacheAge.Round(time.Second))
		}
	}
	
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

	// Apply rate limiting before making the request
	if err := c.checkAndRespectRateLimit(); err != nil {
		if c.verbose {
			log.Printf("ComicVine API Rate Limit Error: %v", err)
		}
		return nil, fmt.Errorf("rate limit: %w", err)
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
	
	// Check for 429 Too Many Requests response
	if resp.StatusCode == 429 {
		// Handle rate limiting response
		resetTimeStr := resp.Header.Get("X-RateLimit-Reset")
		retryAfterStr := resp.Header.Get("Retry-After")
		
		// Default retry after 1 hour if not specified
		retryDelay := 1 * time.Hour
		
		// Try to parse Retry-After header if provided
		if retryAfterStr != "" {
			if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
				retryDelay = time.Duration(seconds) * time.Second
			}
		}
		
		if c.verbose {
			log.Printf("ComicVine API Rate Limit Hit (429 Too Many Requests)")
			log.Printf("  X-RateLimit-Reset: %s", resetTimeStr)
			log.Printf("  Retry-After: %s", retryAfterStr)
			log.Printf("  Will retry after: %v", retryDelay)
		}
		
		// Update our rate limit counter to avoid further requests
		c.requestCount = 200 // Mark as exceeded
		c.requestCountResetTime = time.Now().Add(retryDelay)
		
		return nil, fmt.Errorf("ComicVine API rate limit exceeded. Try again after %v", 
			retryDelay)
	}

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
		// Check for specific API error messages about rate limits
		if strings.Contains(strings.ToLower(result.Error), "rate limit") || 
		   strings.Contains(strings.ToLower(result.Error), "api usage limits") {
			
			// Handle rate limiting error message
			if c.verbose {
				log.Printf("ComicVine API Rate Limit Error: %s", result.Error)
			}
			
			// Update our rate limit counter to avoid further requests
			c.requestCount = 200 // Mark as exceeded
			c.requestCountResetTime = time.Now().Add(1 * time.Hour)
			
			return nil, fmt.Errorf("ComicVine API rate limit exceeded: %s", result.Error)
		}
		
		// General API error
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

	// Store results in cache for future use
	c.cache[cacheKey] = CacheEntry{
		Results:   result.Results,
		Timestamp: time.Now(),
	}
	
	if c.verbose {
		log.Printf("ComicVine API: Stored %d results in cache for '%s' (expires: %v)", 
			len(result.Results), cacheKey, time.Now().Add(24*time.Hour).Format(time.RFC3339))
	}

	return bestMatch, nil
}