package models

import "time"

// ParsedFilename represents the LLM-extracted information from a comic filename
type ParsedFilename struct {
	OriginalFilename string `json:"original_filename"`
	Title            string `json:"title"`
	IssueNumber      string `json:"issue_number"`
	Year             string `json:"year,omitempty"`
	Publisher        string `json:"publisher,omitempty"`
	VolumeNumber     string `json:"volume_number,omitempty"`
	Confidence       string `json:"confidence"` // high, medium, low
	Notes            string `json:"notes,omitempty"`
}

// ComicVineSearchParams holds the parameters for a ComicVine search
type ComicVineSearchParams struct {
	Title       string
	IssueNumber string
}

// ComicVineIssue represents a comic issue from ComicVine API
type ComicVineIssue struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	IssueNumber   string    `json:"issue_number"`
	CoverDate     string    `json:"cover_date"`
	StoreDate     string    `json:"store_date"`
	Description   string    `json:"description"`
	SiteDetailURL string    `json:"site_detail_url"`
	Volume        VolumeRef `json:"volume"`
	Image         ImageRef  `json:"image"`
}

// VolumeRef is a reference to a volume in ComicVine
type VolumeRef struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	SiteURL   string `json:"site_detail_url"`
	Publisher string `json:"publisher_name,omitempty"` // We'll populate this
}

// ImageRef holds image URLs from ComicVine
type ImageRef struct {
	SmallURL  string `json:"small_url"`
	MediumURL string `json:"medium_url"`
	LargeURL  string `json:"large_url"`
}

// ComicVineResponse is the API response wrapper
type ComicVineResponse struct {
	Error                string           `json:"error"`
	Limit                int              `json:"limit"`
	Offset               int              `json:"offset"`
	NumberOfPageResults  int              `json:"number_of_page_results"`
	NumberOfTotalResults int              `json:"number_of_total_results"`
	StatusCode           int              `json:"status_code"`
	Results              []ComicVineIssue `json:"results"`
}

// ComicVineVolumeResponse for volume lookups
type ComicVineVolumeResponse struct {
	Error      string          `json:"error"`
	StatusCode int             `json:"status_code"`
	Results    ComicVineVolume `json:"results"`
}

// ComicVineVolume represents volume details
type ComicVineVolume struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	StartYear string       `json:"start_year"`
	Publisher PublisherRef `json:"publisher"`
}

// PublisherRef is a reference to a publisher
type PublisherRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MatchResult represents the LLM's choice from ComicVine results
type MatchResult struct {
	OriginalFilename string          `json:"original_filename"`
	ParsedInfo       ParsedFilename  `json:"parsed_info"`
	SelectedIssue    *ComicVineIssue `json:"selected_issue,omitempty"`
	MatchConfidence  string          `json:"match_confidence"` // high, medium, low, none
	Reasoning        string          `json:"reasoning"`
	ComicVineID      int             `json:"comicvine_id,omitempty"`
	ComicVineURL     string          `json:"comicvine_url,omitempty"`
}

// ProcessingResult is the final output for each file
type ProcessingResult struct {
	Filename        string       `json:"filename"`
	Success         bool         `json:"success"`
	Error           string       `json:"error,omitempty"`
	Match           *MatchResult `json:"match,omitempty"`
	ProcessedAt     time.Time    `json:"processed_at"`
	ProcessingTimeMS int64       `json:"processing_time_ms"`
}

// BatchProgress tracks progress of batch processing
type BatchProgress struct {
	Total      int `json:"total"`
	Processed  int `json:"processed"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
	Skipped    int `json:"skipped"`
}
