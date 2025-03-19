package comicvine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAPIClient is a mock implementation of the APIClient interface
type MockAPIClient struct {
	mock.Mock
}

// Request is the mock implementation of the APIClient.Request method
func (m *MockAPIClient) Request(ctx context.Context, endpoint string, params map[string]string) ([]byte, int, error) {
	args := m.Called(ctx, endpoint, params)
	return args.Get(0).([]byte), args.Int(1), args.Error(2)
}

// Helper function to load test data files
func loadTestData(t *testing.T, filename string) []byte {
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load test data file %s: %v", path, err)
	}
	return data
}

// TestNewService tests the creation of a new ComicService
func TestNewService(t *testing.T) {
	// Arrange
	apiKey := "test-api-key"
	verbose := true

	// Act
	service := NewService(apiKey, verbose)

	// Assert
	assert.NotNil(t, service, "Service should not be nil")
	assert.NotNil(t, service.client, "Service client should not be nil")
	assert.Equal(t, verbose, service.verbose, "Service verbose setting should match input")
}

// TestSearchSeries tests searching for a comic series
func TestSearchSeries(t *testing.T) {
	// Load test data from files
	successfulResponse := loadTestData(t, "search_volume_response.json")

	// Test cases
	testCases := []struct {
		name           string
		seriesName     string
		seriesYear     string // Optional year parameter
		mockResponse   []byte
		mockStatusCode int
		mockError      error
		expectedError  bool
		expectedCount  int
	}{
		{
			name:           "successful series search",
			seriesName:     "Batman",
			seriesYear:     "",
			mockResponse:   successfulResponse,
			mockStatusCode: 200,
			mockError:      nil,
			expectedError:  false,
			expectedCount:  10, // Adjust based on actual response
		},
		{
			name:           "successful series search with year",
			seriesName:     "Batman",
			seriesYear:     "2020",
			mockResponse:   successfulResponse,
			mockStatusCode: 200,
			mockError:      nil,
			expectedError:  false,
			expectedCount:  10, // Adjust based on actual response
		},
		{
			name:           "api error",
			seriesName:     "ErrorSeries",
			seriesYear:     "",
			mockResponse:   nil,
			mockStatusCode: 500,
			mockError:      fmt.Errorf("API error"),
			expectedError:  true,
			expectedCount:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := new(MockAPIClient)

			// Set up expectations for the mock
			mockClient.On("Request",
				mock.Anything, // context
				"search",      // endpoint
				mock.MatchedBy(func(params map[string]string) bool {
					resourceVal, resourceExists := params["resources"]
					filterVal, filterExists := params["filter"]

					if !resourceExists || resourceVal != "volume" || !filterExists {
						return false
					}

					var expectedFilterVal string
					if tc.seriesYear != "" {
						expectedFilterVal = fmt.Sprintf("name:%s,start_year:%s", tc.seriesName, tc.seriesYear)
					} else {
						expectedFilterVal = "name:" + tc.seriesName
					}

					return filterVal == expectedFilterVal
				}),
			).Return(tc.mockResponse, tc.mockStatusCode, tc.mockError)

			service := &ComicService{
				client:  mockClient,
				verbose: false,
			}

			// Act
			results, err := service.searchSeries(context.Background(), tc.seriesName, tc.seriesYear)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Did not expect an error")
				assert.Equal(t, tc.expectedCount, len(results), "Expected count of results does not match")

				// If we expect results, let's do some deeper validation of the first result
				if tc.expectedCount > 0 {
					// These assertions should match what's in your test data file
					// Adjust these based on the actual content of your test file
					assert.NotEmpty(t, results[0].Name, "Series name should not be empty")
					assert.NotZero(t, results[0].ComicVineID, "Series ID should not be zero")
				}
			}

			// Verify all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}

// TestGetIssue tests retrieving a specific issue by volume ID and issue number
func TestGetIssue(t *testing.T) {
	// Load test data from files
	successfulResponse := loadTestData(t, "issue_response.json")

	// Test cases
	testCases := []struct {
		name            string
		volumeID        int
		issueNumber     string
		mockResponse    []byte
		mockStatusCode  int
		mockError       error
		expectedError   bool
		expectedIssueID int
	}{
		{
			name:            "successful issue retrieval",
			volumeID:        1234, // Adjust to match what's in your test data
			issueNumber:     "1",  // Adjust to match what's in your test data
			mockResponse:    successfulResponse,
			mockStatusCode:  200,
			mockError:       nil,
			expectedError:   false,
			expectedIssueID: 5678, // Adjust to match what's in your test data
		},
		{
			name:            "issue not found",
			volumeID:        1234,
			issueNumber:     "999",
			mockResponse:    []byte(`{"status_code":1,"results":{"results":[]}}`),
			mockStatusCode:  200,
			mockError:       nil,
			expectedError:   true,
			expectedIssueID: 0,
		},
		{
			name:            "api error",
			volumeID:        1234,
			issueNumber:     "1",
			mockResponse:    nil,
			mockStatusCode:  500,
			mockError:       fmt.Errorf("API error"),
			expectedError:   true,
			expectedIssueID: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := new(MockAPIClient)

			// Set up expectations
			mockClient.On("Request",
				mock.Anything, // context
				"search",      // endpoint
				mock.MatchedBy(func(params map[string]string) bool {
					resourceVal, resourceExists := params["resources"]
					filterVal, filterExists := params["filter"]

					if !resourceExists || resourceVal != "volume" || !filterExists {
						return false
					}

					var expectedFilterVal string
					if tc.volumeID != 1234 {
						expectedFilterVal = fmt.Sprintf("volumeID:%d", tc.volumeID)
					}

					return filterVal == expectedFilterVal
				}),
			).Return(tc.mockResponse, tc.mockStatusCode, tc.mockError)
			service := &ComicService{
				client:  mockClient,
				verbose: false,
			}

			// Act
			result, err := service.getIssue(context.Background(), tc.volumeID, tc.issueNumber)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
				assert.Nil(t, result, "Result should be nil when there's an error")
			} else {
				assert.NoError(t, err, "Did not expect an error")
				assert.NotNil(t, result, "Result should not be nil")

				// Validate specific fields based on your test data
				// These assertions should match what's in your actual test data file
				assert.Equal(t, tc.expectedIssueID, result[0].ComicVineID, "Issue ID does not match")
				assert.Equal(t, tc.issueNumber, result[0].IssueNumber, "Issue number does not match")

				// Add more specific assertions based on the actual content of your test data
				assert.NotEmpty(t, result[0].VolumeID, "Volume name should not be empty")
			}

			// Verify all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}

/*
// TestGetComicInfo tests the high-level method to get comic information
// This is the method that implements your described workflow
func TestGetComicInfo(t *testing.T) {
	// Load test data from files
	successfulSeriesResponse := loadTestData(t, "search_volume_response.json")
	successfulIssueResponse := loadTestData(t, "search_issue_response.json")

	// Test cases
	testCases := []struct {
		name            string
		seriesName      string
		issueNumber     string
		seriesYear      string
		mockSeriesResp  []byte
		mockIssueResp   []byte
		mockStatusCodes []int
		mockErrors      []error
		expectedError   bool
	}{
		{
			name:            "successful end-to-end retrieval",
			seriesName:      "Batman",
			issueNumber:     "1",
			seriesYear:      "",
			mockSeriesResp:  successfulSeriesResponse,
			mockIssueResp:   successfulIssueResponse,
			mockStatusCodes: []int{200, 200},
			mockErrors:      []error{nil, nil},
			expectedError:   false,
		},
		{
			name:            "series not found",
			seriesName:      "NonExistentSeries",
			issueNumber:     "1",
			seriesYear:      "",
			mockSeriesResp:  []byte(`{"status_code":1,"results":{"results":[]}}`),
			mockIssueResp:   nil, // Won't be called
			mockStatusCodes: []int{200},
			mockErrors:      []error{nil},
			expectedError:   true,
		},
		{
			name:            "series found but issue not found",
			seriesName:      "Batman",
			issueNumber:     "999",
			seriesYear:      "",
			mockSeriesResp:  successfulSeriesResponse,
			mockIssueResp:   []byte(`{"status_code":1,"results":{"results":[]}}`),
			mockStatusCodes: []int{200, 200},
			mockErrors:      []error{nil, nil},
			expectedError:   true,
		},
		{
			name:            "series search error",
			seriesName:      "ErrorSeries",
			issueNumber:     "1",
			seriesYear:      "",
			mockSeriesResp:  nil,
			mockIssueResp:   nil, // Won't be called
			mockStatusCodes: []int{500},
			mockErrors:      []error{fmt.Errorf("API error")},
			expectedError:   true,
		},
		{
			name:            "issue search error",
			seriesName:      "Batman",
			issueNumber:     "1",
			seriesYear:      "",
			mockSeriesResp:  successfulSeriesResponse,
			mockIssueResp:   nil,
			mockStatusCodes: []int{200, 500},
			mockErrors:      []error{nil, fmt.Errorf("API error")},
			expectedError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockClient := new(MockAPIClient)

			// Extract volume ID from your test data for the issue request matcher
			// This would need to be adjusted based on your actual test data structure
			volumeID := 1234 // Replace with actual ID from your test data

			// Set up expectations for series search
			if tc.mockSeriesResp != nil {
				mockClient.On("Request",
					mock.Anything, // context
					"search",      // endpoint
					mock.MatchedBy(func(params map[string]string) bool {
						resourceVal, resourceExists := params["resources"]
						filterVal, filterExists := params["filter"]

						if !resourceExists || resourceVal != "volume" || !filterExists {
							return false
						}

						var expectedFilterVal string
						if tc.seriesYear != "" {
							expectedFilterVal = fmt.Sprintf("name:%s,start_year:%s", tc.seriesName, tc.seriesYear)
						} else {
							expectedFilterVal = "name:" + tc.seriesName
						}

						return filterVal == expectedFilterVal
					}),
				).Return(tc.mockSeriesResp, tc.mockStatusCodes[0], tc.mockErrors[0])
			}

			// Set up expectations for issue search if needed
			if tc.mockIssueResp != nil {
				mockClient.On("Request",
					mock.Anything, // context
					"issues",      // endpoint
					mock.MatchedBy(func(params map[string]string) bool {
						filterVal, filterExists := params["filter"]

						if !filterExists {
							return false
						}

						expectedFilterVal := fmt.Sprintf("volume:%d,issue_number:%s", volumeID, tc.issueNumber)

						return filterVal == expectedFilterVal
					}),
				).Return(tc.mockIssueResp, tc.mockStatusCodes[1], tc.mockErrors[1])
			}

			service := &ComicService{
				client:  mockClient,
				verbose: false,
			}

			// Act
			result, err := service.GetComicInfo(context.Background(), tc.seriesName, tc.issueNumber, tc.seriesYear)

			// Assert
			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Did not expect an error")
				assert.NotNil(t, result, "Result should not be nil")

				// Add specific assertions based on your actual test data
				assert.Equal(t, tc.seriesName, result.Series, "Series name doesn't match")
				assert.Equal(t, tc.issueNumber, result.Issue, "Issue number doesn't match")
			}

			// Verify all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}

// TestRateLimitHandling tests how the service handles rate limit responses
func TestRateLimitHandling(t *testing.T) {
	// You would need a rate limit response example file
	// This is a hypothetical test for when you get such an example

	// Load test data - you'll need to create this file
	// rateLimitResponse := loadTestData(t, "rate_limit_response.json")

	// For now, let's use a mocked rate limit response
	rateLimitResponse := []byte(`{
		"status_code": 107,
		"error": "API usage limit exceeded. Please try again later.",
		"results": {}
	}`)

	// Arrange
	mockClient := new(MockAPIClient)

	// Set up expectations
	mockClient.On("Request",
		mock.Anything, // context
		"search",      // endpoint
		mock.Anything, // params
	).Return(rateLimitResponse, 429, nil) // 429 is Too Many Requests

	service := &ComicService{
		client:  mockClient,
		verbose: false,
	}

	// Act
	results, err := service.SearchSeries(context.Background(), "Batman", "")

	// Assert
	assert.Error(t, err, "Expected a rate limit error")
	assert.Nil(t, results, "Results should be nil when rate limited")

	// Check if the error message indicates a rate limit issue
	assert.Contains(t, err.Error(), "rate limit", "Error should mention rate limiting")

	// Verify all expectations were met
	mockClient.AssertExpectations(t)
}

// TestInvalidApiKeyHandling tests how the service handles invalid API key responses
func TestInvalidApiKeyHandling(t *testing.T) {
	// You would need an invalid API key response example file
	// This is a hypothetical test for when you get such an example

	// Load test data - you'll need to create this file
	// invalidKeyResponse := loadTestData(t, "invalid_api_key_response.json")

	// For now, let's use a mocked invalid API key response
	invalidKeyResponse := []byte(`{
		"status_code": 100,
		"error": "Invalid API key",
		"results": {}
	}`)

	// Arrange
	mockClient := new(MockAPIClient)

	// Set up expectations
	mockClient.On("Request",
		mock.Anything, // context
		"search",      // endpoint
		mock.Anything, // params
	).Return(invalidKeyResponse, 401, nil) // 401 is Unauthorized

	service := &ComicService{
		client:  mockClient,
		verbose: false,
	}

	// Act
	results, err := service.SearchSeries(context.Background(), "Batman", "")

	// Assert
	assert.Error(t, err, "Expected an API key error")
	assert.Nil(t, results, "Results should be nil when API key is invalid")

	// Check if the error message indicates an API key issue
	assert.Contains(t, err.Error(), "API key", "Error should mention API key")

	// Verify all expectations were met
	mockClient.AssertExpectations(t)
}
*/
