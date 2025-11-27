package comicvine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"comic-parser/config"
	"comic-parser/models"
)

// Client is a ComicVine API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	// Rate limiting
	rateLimiter *time.Ticker
	rateMutex   sync.Mutex

	// Volume cache to reduce API calls
	volumeCache map[int]*models.ComicVineVolume
	cacheMutex  sync.RWMutex
}

// NewClient creates a new ComicVine API client
func NewClient(cfg *config.Config) *Client {
	// ComicVine has a rate limit, default to ~1 request per second
	ratePerSecond := cfg.RateLimitPerMin / 60
	if ratePerSecond < 1 {
		ratePerSecond = 1
	}

	return &Client{
		apiKey:  cfg.ComicVineAPIKey,
		baseURL: cfg.ComicVineAPIBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: time.NewTicker(time.Second / time.Duration(ratePerSecond)),
		volumeCache: make(map[int]*models.ComicVineVolume),
	}
}

// SearchIssues searches for comic issues by title and optional issue number
func (c *Client) SearchIssues(ctx context.Context, title string, issueNumber string) ([]models.ComicVineIssue, error) {
	// Respect rate limit
	c.rateMutex.Lock()
	<-c.rateLimiter.C
	c.rateMutex.Unlock()

	// Build search query
	// ComicVine's search is best when searching for volumes first
	// then filtering by issue number
	issues, err := c.searchByVolumeAndIssue(ctx, title, issueNumber)
	if err != nil {
		return nil, err
	}

	// Enrich results with publisher info
	for i := range issues {
		if issues[i].Volume.ID > 0 {
			vol, err := c.getVolume(ctx, issues[i].Volume.ID)
			if err == nil && vol != nil && vol.Publisher.Name != "" {
				issues[i].Volume.Publisher = vol.Publisher.Name
			}
		}
	}

	return issues, nil
}

// searchByVolumeAndIssue performs a search using the issues endpoint with filters
func (c *Client) searchByVolumeAndIssue(ctx context.Context, title string, issueNumber string) ([]models.ComicVineIssue, error) {
	// First, search for the volume
	volumes, err := c.searchVolumes(ctx, title)
	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 {
		// Fall back to general issue search
		return c.searchIssuesDirectly(ctx, title, issueNumber)
	}

	// Collect volume IDs for filtering
	var allIssues []models.ComicVineIssue
	seen := make(map[int]bool)

	// Check top matching volumes for the issue
	maxVolumesToCheck := 5
	if len(volumes) < maxVolumesToCheck {
		maxVolumesToCheck = len(volumes)
	}

	for _, vol := range volumes[:maxVolumesToCheck] {
		issues, err := c.getIssuesForVolume(ctx, vol.ID, issueNumber)
		if err != nil {
			continue // Don't fail entirely if one volume lookup fails
		}

		for _, issue := range issues {
			if !seen[issue.ID] {
				seen[issue.ID] = true
				// Add volume info
				issue.Volume = models.VolumeRef{
					ID:   vol.ID,
					Name: vol.Name,
				}
				if vol.Publisher.Name != "" {
					issue.Volume.Publisher = vol.Publisher.Name
				}
				allIssues = append(allIssues, issue)
			}
		}
	}

	if len(allIssues) == 0 {
		// Fall back to direct search
		return c.searchIssuesDirectly(ctx, title, issueNumber)
	}

	return allIssues, nil
}

// searchVolumes searches for volumes (comic series) by name
func (c *Client) searchVolumes(ctx context.Context, name string) ([]models.ComicVineVolume, error) {
	// Respect rate limit
	c.rateMutex.Lock()
	<-c.rateLimiter.C
	c.rateMutex.Unlock()

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("format", "json")
	params.Set("resources", "volume")
	params.Set("query", name)
	params.Set("limit", "10")
	params.Set("field_list", "id,name,start_year,publisher")

	reqURL := fmt.Sprintf("%s/search/?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "ComicParser/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Results []models.ComicVineVolume `json:"results"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result.Results, nil
}

// getIssuesForVolume gets issues for a specific volume, optionally filtered by issue number
func (c *Client) getIssuesForVolume(ctx context.Context, volumeID int, issueNumber string) ([]models.ComicVineIssue, error) {
	// Respect rate limit
	c.rateMutex.Lock()
	<-c.rateLimiter.C
	c.rateMutex.Unlock()

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("format", "json")
	params.Set("limit", "100")
	params.Set("field_list", "id,name,issue_number,cover_date,store_date,site_detail_url,volume,image")

	// Filter by volume
	filter := fmt.Sprintf("volume:%d", volumeID)
	if issueNumber != "" {
		// Normalize issue number for comparison
		normalizedIssue := normalizeIssueNumber(issueNumber)
		filter += fmt.Sprintf(",issue_number:%s", normalizedIssue)
	}
	params.Set("filter", filter)

	reqURL := fmt.Sprintf("%s/issues/?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "ComicParser/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result models.ComicVineResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result.Results, nil
}

// searchIssuesDirectly searches issues directly (fallback method)
func (c *Client) searchIssuesDirectly(ctx context.Context, title string, issueNumber string) ([]models.ComicVineIssue, error) {
	// Respect rate limit
	c.rateMutex.Lock()
	<-c.rateLimiter.C
	c.rateMutex.Unlock()

	// Build search query
	query := title
	if issueNumber != "" {
		query += " " + issueNumber
	}

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("format", "json")
	params.Set("resources", "issue")
	params.Set("query", query)
	params.Set("limit", "10")
	params.Set("field_list", "id,name,issue_number,cover_date,store_date,site_detail_url,volume,image")

	reqURL := fmt.Sprintf("%s/search/?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "ComicParser/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result models.ComicVineResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result.Results, nil
}

// getVolume retrieves volume details (with caching)
func (c *Client) getVolume(ctx context.Context, volumeID int) (*models.ComicVineVolume, error) {
	// Check cache first
	c.cacheMutex.RLock()
	if vol, ok := c.volumeCache[volumeID]; ok {
		c.cacheMutex.RUnlock()
		return vol, nil
	}
	c.cacheMutex.RUnlock()

	// Respect rate limit
	c.rateMutex.Lock()
	<-c.rateLimiter.C
	c.rateMutex.Unlock()

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("format", "json")
	params.Set("field_list", "id,name,start_year,publisher")

	reqURL := fmt.Sprintf("%s/volume/4050-%d/?%s", c.baseURL, volumeID, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "ComicParser/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result models.ComicVineVolumeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	// Cache the result
	c.cacheMutex.Lock()
	c.volumeCache[volumeID] = &result.Results
	c.cacheMutex.Unlock()

	return &result.Results, nil
}

// normalizeIssueNumber removes leading zeros and normalizes issue numbers
func normalizeIssueNumber(issue string) string {
	issue = strings.TrimSpace(issue)
	issue = strings.TrimPrefix(issue, "#")
	issue = strings.TrimLeft(issue, "0")
	if issue == "" {
		return "0"
	}
	return issue
}

// Close cleans up the client resources
func (c *Client) Close() {
	if c.rateLimiter != nil {
		c.rateLimiter.Stop()
	}
}
