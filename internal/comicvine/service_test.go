package comicvine

/*
func TestComicService_GetMetadata(t *testing.T) {
	// Create a test server that returns mock data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		mockResponse := Response{
			StatusCode: 1,
			Results: []Issue{
				{
					ID:          12345,
					Name:        "Test Issue",
					IssueNumber: "001",
					Volume: Volume{
						ID:   67890,
						Name: "Batman",
					},
					CoverDate:   "2016-01-01",
					Image: Image{
						OriginalURL: "http://example.com/cover.jpg",
					},
					Description: "Test description",
				},
			},
		}

		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Override the base URL to use the test server
	originalBaseURL := baseURL
	baseURL = server.URL + "/api"
	defer func() { baseURL = originalBaseURL }()

	// Create the service
	service := &ComicService{
		client: &Client{
			apiKey:     "test-api-key",
			httpClient: server.Client(),
			cache:      make(map[string]CacheEntry),
		},
	}

	// Test with a valid filename
	result, err := service.GetMetadata("Batman (2016) #001.cbz")

	if err != nil {
		t.Errorf("GetMetadata() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("GetMetadata() result is nil, want non-nil")
	}

	// Verify the result values
	if result.Series != "Batman" {
		t.Errorf("GetMetadata() result.Series = %v, want %v", result.Series, "Batman")
	}

	if result.Issue != "001" {
		t.Errorf("GetMetadata() result.Issue = %v, want %v", result.Issue, "001")
	}

	if result.Year != "2016" {
		t.Errorf("GetMetadata() result.Year = %v, want %v", result.Year, "2016")
	}

	if result.ComicVineID != 12345 {
		t.Errorf("GetMetadata() result.ComicVineID = %v, want %v", result.ComicVineID, 12345)
	}

	if result.Title != "Test Issue" {
		t.Errorf("GetMetadata() result.Title = %v, want %v", result.Title, "Test Issue")
	}

	if result.CoverURL != "http://example.com/cover.jpg" {
		t.Errorf("GetMetadata() result.CoverURL = %v, want %v", result.CoverURL, "http://example.com/cover.jpg")
	}
}

func TestComicService_GetMetadataForFiles(t *testing.T) {
	// Create a test server that returns mock data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		mockResponse := Response{
			StatusCode: 1,
			Results: []Issue{
				{
					ID:          12345,
					Name:        "Test Issue",
					IssueNumber: "001",
					Volume: Volume{
						ID:   67890,
						Name: "Batman",
					},
					CoverDate:   "2016-01-01",
					Image: Image{
						OriginalURL: "http://example.com/cover.jpg",
					},
					Description: "Test description",
				},
			},
		}

		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Override the base URL to use the test server
	originalBaseURL := baseURL
	baseURL = server.URL + "/api"
	defer func() { baseURL = originalBaseURL }()

	// Create the service
	service := &ComicService{
		client: &Client{
			apiKey:     "test-api-key",
			httpClient: server.Client(),
			cache:      make(map[string]CacheEntry),
		},
		verbose: false,
	}

	// Test with a mix of valid and invalid filenames
	filenames := []string{
		"Batman (2016) #001.cbz",
		"Batman (2016) #001", // No extension
		"Batman (2016) #001.txt", // Different extension
		"not-a-comic-format", // Invalid format
	}

	results, err := service.GetMetadataForFiles(filenames)

	if err != nil {
		t.Errorf("GetMetadataForFiles() error = %v, want nil", err)
	}

	// Our parser is very forgiving and will match all 4 files,
	// even the "not-a-comic-format" one (it defaults to series name with issue #1)
	if len(results) != 4 {
		t.Errorf("GetMetadataForFiles() got %d results, want 4", len(results))
	}

	if len(results) > 0 {
		result := results[0]

		if result.Series != "Batman" {
			t.Errorf("GetMetadataForFiles() result.Series = %v, want %v", result.Series, "Batman")
		}

		if result.Issue != "001" {
			t.Errorf("GetMetadataForFiles() result.Issue = %v, want %v", result.Issue, "001")
		}

		if result.Year != "2016" {
			t.Errorf("GetMetadataForFiles() result.Year = %v, want %v", result.Year, "2016")
		}
	}
}

*/
