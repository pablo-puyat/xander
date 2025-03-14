package comicvine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey, false)

	if client.apiKey != apiKey {
		t.Errorf("NewClient() apiKey = %v, want %v", client.apiKey, apiKey)
	}

	if client.httpClient == nil {
		t.Error("NewClient() httpClient is nil, want non-nil")
	}
}

func TestGetIssue(t *testing.T) {
	// Create a test server that returns a mock response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.URL.Path != "/api/issues" {
			t.Errorf("Expected URL path to be /api/issues, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("api_key") != "test-api-key" {
			t.Errorf("Expected api_key to be test-api-key, got %s", query.Get("api_key"))
		}

		if query.Get("format") != "json" {
			t.Errorf("Expected format to be json, got %s", query.Get("format"))
		}

		// Return a mock response
		mockResponse := Response{
			StatusCode: 1,
			Results: []Issue{
				{
					ID:          12345,
					Name:        "Test Issue",
					IssueNumber: "1",
					Volume: Volume{
						ID:   67890,
						Name: "Test Series",
					},
					CoverDate: "2020-01-01",
					Image: Image{
						OriginalURL: "http://example.com/cover.jpg",
					},
					Description: "Test description",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create a client that uses the test server
	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
		cache:      make(map[string]CacheEntry),
	}
	// Override the base URL to use the test server
	originalBaseURL := baseURL
	baseURL = server.URL + "/api"
	defer func() { baseURL = originalBaseURL }()

	// Call the method under test
	issue, err := client.GetIssue("Test Series", "1")

	// Verify the result
	if err != nil {
		t.Errorf("GetIssue() error = %v, want nil", err)
	}

	if issue == nil {
		t.Fatal("GetIssue() issue is nil, want non-nil")
	}

	if issue.ID != 12345 {
		t.Errorf("GetIssue() issue.ID = %v, want %v", issue.ID, 12345)
	}

	if issue.Name != "Test Issue" {
		t.Errorf("GetIssue() issue.Name = %v, want %v", issue.Name, "Test Issue")
	}

	if issue.IssueNumber != "1" {
		t.Errorf("GetIssue() issue.IssueNumber = %v, want %v", issue.IssueNumber, "1")
	}

	if issue.Volume.Name != "Test Series" {
		t.Errorf("GetIssue() issue.Volume.Name = %v, want %v", issue.Volume.Name, "Test Series")
	}
}

func TestGetIssue_Error(t *testing.T) {
	// Create a test server that returns an error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a mock error response
		mockResponse := Response{
			StatusCode: 100,
			Error:      "API Error",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create a client that uses the test server
	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
		cache:      make(map[string]CacheEntry),
	}
	// Override the base URL to use the test server
	originalBaseURL := baseURL
	baseURL = server.URL + "/api"
	defer func() { baseURL = originalBaseURL }()

	// Call the method under test
	issue, err := client.GetIssue("Test Series", "1")

	// Verify the result
	if err == nil {
		t.Error("GetIssue() error is nil, want non-nil")
	}

	if issue != nil {
		t.Errorf("GetIssue() issue = %v, want nil", issue)
	}
}

func TestGetIssue_NoResults(t *testing.T) {
	// Create a test server that returns an empty result
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a mock response with no results
		mockResponse := Response{
			StatusCode: 1,
			Results:    []Issue{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create a client that uses the test server
	client := &Client{
		apiKey:     "test-api-key",
		httpClient: server.Client(),
		cache:      make(map[string]CacheEntry),
	}
	// Override the base URL to use the test server
	originalBaseURL := baseURL
	baseURL = server.URL + "/api"
	defer func() { baseURL = originalBaseURL }()

	// Call the method under test
	issue, err := client.GetIssue("Test Series", "1")

	// Verify the result
	if err == nil {
		t.Error("GetIssue() error is nil, want non-nil")
	}

	if issue != nil {
		t.Errorf("GetIssue() issue = %v, want nil", issue)
	}
}