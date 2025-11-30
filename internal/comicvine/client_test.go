package comicvine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"comic-parser/internal/config"
	"comic-parser/internal/models"
)

func TestNormalizeIssueNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1", "1"},
		{"01", "1"},
		{"001", "1"},
		{"#1", "1"},
		{"#001", "1"},
		{"1.1", "1.1"},
		{"  1  ", "1"},
		{"0", "0"},
		{"", "0"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Input_%s", tt.input), func(t *testing.T) {
			got := normalizeIssueNumber(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeIssueNumber(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		ComicVineAPIKey:     "test-key",
		ComicVineAPIBaseURL: "http://example.com",
	}

	client := NewClient(cfg, http.DefaultClient)
	defer client.Close()

	if client.apiKey != "test-key" {
		t.Errorf("NewClient().apiKey = %s; want test-key", client.apiKey)
	}
	if client.baseURL != "http://example.com" {
		t.Errorf("NewClient().baseURL = %s; want http://example.com", client.baseURL)
	}
	if client.httpClient == nil {
		t.Error("NewClient().httpClient is nil")
	}
	if client.volumeCache == nil {
		t.Error("NewClient().volumeCache is nil")
	}
}

func TestSearchVolumes(t *testing.T) {
	// Mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/search/" {
			t.Errorf("Expected path /search/, got %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("api_key") != "test-key" {
			t.Errorf("Expected api_key test-key, got %s", query.Get("api_key"))
		}
		if query.Get("query") != "Test Volume" {
			t.Errorf("Expected query 'Test Volume', got %s", query.Get("query"))
		}
		if query.Get("resources") != "volume" {
			t.Errorf("Expected resources volume, got %s", query.Get("resources"))
		}

		// Mock response
		resp := struct {
			Results []models.ComicVineVolume `json:"results"`
		}{
			Results: []models.ComicVineVolume{
				{ID: 100, Name: "Test Volume", StartYear: "2020"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	cfg := &config.Config{
		ComicVineAPIKey:     "test-key",
		ComicVineAPIBaseURL: ts.URL,
	}

	client := NewClient(cfg, ts.Client())
	defer client.Close()

	// Speed up rate limiter for test
	client.rateLimiter.Stop()
	client.rateLimiter = time.NewTicker(1 * time.Millisecond)

	ctx := context.Background()
	results, err := client.searchVolumes(ctx, "Test Volume")
	if err != nil {
		t.Fatalf("searchVolumes failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Test Volume" {
		t.Errorf("Expected volume name 'Test Volume', got %s", results[0].Name)
	}
}
