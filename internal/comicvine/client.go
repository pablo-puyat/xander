package comicvine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL         = "https://comicvine.gamespot.com/api"
	defaultRequestInterval = 1 * time.Second // Minimum interval between requests to prevent API velocity detection
	maxRequestsPerHour     = 200             // Maximum allowed requests per hour by ComicVine API
)

type Client struct {
	apiKey             string
	baseURL            string
	httpClient         *http.Client
	verbose            bool
	lastRequestTime    time.Time
	requestCount       int
	requestCountReset  time.Time
	minRequestInterval time.Duration
}

func NewClient(apiKey string, verbose bool) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		verbose:            verbose,
		lastRequestTime:    time.Time{}, // Zero time
		requestCount:       0,
		requestCountReset:  time.Now().Add(time.Hour),
		minRequestInterval: defaultRequestInterval,
	}, nil
}

// APIResponse represents the standard ComicVine API response structure
type APIResponse struct {
	StatusCode int             `json:"status_code"`
	Error      string          `json:"error"`
	Results    json.RawMessage `json:"results"`
}

// Get performs an HTTP GET request to the ComicVine API
func (c *Client) Get(ctx context.Context, endpoint string, params map[string]string) ([]byte, int, error) {
	// Check rate limits before making request
	if err := c.checkRateLimit(); err != nil {
		return nil, 0, err
	}

	// Construct the request URL
	requestURL, err := c.buildRequestURL(endpoint, params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build request URL: %w", err)
	}

	// Log the request URL (with API key masked) if verbose mode is enabled
	if c.verbose {
		maskedURL := c.getMaskedURL(requestURL)
		log.Printf("ComicVine API Request: GET %s", maskedURL)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute the HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Track this request for rate limiting
	c.updateRequestTracking()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	// Log the response if verbose mode is enabled
	if c.verbose {
		log.Printf("ComicVine API Response Status: %s", resp.Status)
		if len(body) > 1000 {
			log.Printf("ComicVine API Response (truncated): %s...", body[:1000])
		} else {
			log.Printf("ComicVine API Response: %s", body)
		}
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		// Handle rate limiting specifically
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := c.parseRetryAfterHeader(resp)

			// Update our rate limit tracking
			c.requestCount = maxRequestsPerHour // Mark as exceeded
			c.requestCountReset = time.Now().Add(retryAfter)

			if c.verbose {
				log.Printf("ComicVine API Rate Limit Hit: Retry after %v", retryAfter)
			}

			return nil, resp.StatusCode, fmt.Errorf("rate limit exceeded, retry after %v", retryAfter)
		}

		return nil, resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Check if the response is a valid JSON
	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Check for API errors
	if apiResponse.StatusCode != 1 {
		// Check for rate limit specific error messages
		if c.isRateLimitError(apiResponse.Error) {
			// Update rate limit tracking
			c.requestCount = maxRequestsPerHour
			c.requestCountReset = time.Now().Add(time.Hour)

			return nil, resp.StatusCode, fmt.Errorf("API rate limit exceeded: %s", apiResponse.Error)
		}

		return nil, resp.StatusCode, fmt.Errorf("API error: %s", apiResponse.Error)
	}

	return body, resp.StatusCode, nil
}

// buildRequestURL constructs the full URL for the API request
func (c *Client) buildRequestURL(endpoint string, params map[string]string) (string, error) {
	// Create the base URL
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	// Add the endpoint to the path
	u.Path += "/" + endpoint

	// Create query parameters
	q := u.Query()

	// Add API key and format parameters
	q.Add("api_key", c.apiKey)
	q.Add("format", "json")

	// Add any additional parameters
	for key, value := range params {
		q.Add(key, value)
	}

	// Encode the parameters back to the URL
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// getMaskedURL returns the URL with the API key masked for logging
func (c *Client) getMaskedURL(originalURL string) string {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		// If parsing fails, mask the key manually with a simple replace
		return strings.Replace(originalURL, c.apiKey, "********", -1)
	}

	q := parsedURL.Query()
	if q.Get("api_key") != "" {
		q.Set("api_key", "********")
		parsedURL.RawQuery = q.Encode()
		return parsedURL.String()
	}

	return originalURL
}

// checkRateLimit checks if we've hit the rate limit and should pause
func (c *Client) checkRateLimit() error {
	now := time.Now()

	// Check if we need to reset the counter (hourly)
	if now.After(c.requestCountReset) {
		if c.verbose {
			log.Printf("ComicVine API: Resetting rate limit counter (hourly reset)")
		}
		c.requestCount = 0
		c.requestCountReset = now.Add(time.Hour)
	}

	// Check if we've exceeded hourly limit
	if c.requestCount >= maxRequestsPerHour {
		resetTime := c.requestCountReset
		waitDuration := resetTime.Sub(now)

		if c.verbose {
			log.Printf("ComicVine API: Rate limit reached (%d/hour). Need to wait %v until %v",
				maxRequestsPerHour, waitDuration.Round(time.Second), resetTime.Format(time.RFC3339))
		}

		return fmt.Errorf("ComicVine API rate limit exceeded (%d/hour). Try again after %v",
			maxRequestsPerHour, resetTime.Format(time.RFC3339))
	}

	// Add delay between requests to prevent "velocity detection" blocks
	if !c.lastRequestTime.IsZero() {
		sinceLastRequest := now.Sub(c.lastRequestTime)
		if sinceLastRequest < c.minRequestInterval {
			sleepDuration := c.minRequestInterval - sinceLastRequest
			if c.verbose {
				log.Printf("ComicVine API: Adding delay of %v between requests",
					sleepDuration.Round(time.Millisecond))
			}
			time.Sleep(sleepDuration)
		}
	}

	return nil
}

// updateRequestTracking updates the request tracking information
func (c *Client) updateRequestTracking() {
	c.lastRequestTime = time.Now()
	c.requestCount++

	if c.verbose {
		log.Printf("ComicVine API: Request %d of %d for this hour (resets at %v)",
			c.requestCount, maxRequestsPerHour, c.requestCountReset.Format(time.RFC3339))
	}
}

// parseRetryAfterHeader parses the Retry-After header from a rate-limited response
func (c *Client) parseRetryAfterHeader(resp *http.Response) time.Duration {
	retryAfterStr := resp.Header.Get("Retry-After")

	// Default retry after 1 hour if not specified
	retryDelay := 1 * time.Hour

	// Try to parse Retry-After header if provided
	if retryAfterStr != "" {
		if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
			retryDelay = time.Duration(seconds) * time.Second
		}
	}

	return retryDelay
}

// isRateLimitError checks if an error message indicates a rate limit was exceeded
func (c *Client) isRateLimitError(errorMsg string) bool {
	errorLower := strings.ToLower(errorMsg)
	return strings.Contains(errorLower, "rate limit") ||
		strings.Contains(errorLower, "api usage limits") ||
		strings.Contains(errorLower, "too many requests")
}
