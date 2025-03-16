package comicvine

import (
	"time"
)

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
	StoreDate   string `json:"store_date"`
	Image       Image  `json:"image"`
	Description string `json:"description"`
	Publisher   string `json:"publisher"`

	// Additional data captured from raw JSON response
	Characters  []map[string]interface{} `json:"-"`
	Teams       []map[string]interface{} `json:"-"`
	Locations   []map[string]interface{} `json:"-"`
	Concepts    []map[string]interface{} `json:"-"`
	Objects     []map[string]interface{} `json:"-"`
	People      []map[string]interface{} `json:"-"`
	DateAdded   string                   `json:"-"`
	DateUpdated string                   `json:"-"`

	// Raw data map for additional fields not explicitly defined
	RawData map[string]interface{} `json:"-"`
}

// Response represents the API response from ComicVine
type Response struct {
	StatusCode int     `json:"status_code"`
	Error      string  `json:"error"`
	Results    []Issue `json:"results"`
}

// CacheEntry represents a cached API response
type CacheEntry struct {
	Results   []Issue
	Timestamp time.Time
}