package comicvine

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockClient implements the ComicVineClient interface for testing
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Get(endpoint string, params map[string]string) ([]byte, int, error) {
	args := m.Called(endpoint, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]byte), args.Int(1), args.Error(2)
}

func TestNewService(t *testing.T) {
	t.Run("creates service with valid API key", func(t *testing.T) {
		service := NewService("test-api-key", false)
		assert.NotNil(t, service)
		assert.False(t, service.verbose)
	})

	t.Run("creates verbose service", func(t *testing.T) {
		service := NewService("test-api-key", true)
		assert.NotNil(t, service)
		assert.True(t, service.verbose)
	})

	t.Run("initializes cache", func(t *testing.T) {
		service := NewService("test-api-key", false)
		assert.NotNil(t, service.cache)
	})
}

func TestService_GetIssue(t *testing.T) {
	tests := []struct {
		name              string
		series            string
		issueNumber       string
		mockResponses     []mockResponse
		expectedIssue     *Issue
		expectedError     string
		expectCacheCheck  bool
		expectAPICall     bool
		expectedCacheSize int
	}{
		{
			name:        "retrieves issue from API successfully",
			series:      "Batman",
			issueNumber: "1",
			mockResponses: []mockResponse{
				{
					endpoint: "issues",
					params: map[string]string{
						"query": "Batman 1",
						"limit": "10",
						"sort":  "name:asc",
					},
					response: responseWithIssues([]Issue{
						{
							ID:          12345,
							Name:        "Batman #1",
							IssueNumber: "1",
							Volume: Volume{
								ID:   67890,
								Name: "Batman",
							},
						},
					}),
					statusCode: 200,
					err:        nil,
				},
			},
			expectedIssue: &Issue{
				ID:          12345,
				Name:        "Batman #1",
				IssueNumber: "1",
				Volume: Volume{
					ID:   67890,
					Name: "Batman",
				},
			},
			expectedError:     "",
			expectCacheCheck:  true,
			expectAPICall:     true,
			expectedCacheSize: 1,
		},
		{
			name:        "normalizes issue number by removing leading zeros",
			series:      "Batman",
			issueNumber: "001",
			mockResponses: []mockResponse{
				{
					endpoint: "issues",
					params: map[string]string{
						"query": "Batman 1", // Should normalize to 1 not 001
						"limit": "10",
						"sort":  "name:asc",
					},
					response: responseWithIssues([]Issue{
						{
							ID:          12345,
							Name:        "Batman #1",
							IssueNumber: "1",
							Volume: Volume{
								ID:   67890,
								Name: "Batman",
							},
						},
					}),
					statusCode: 200,
					err:        nil,
				},
			},
			expectedIssue: &Issue{
				ID:          12345,
				Name:        "Batman #1",
				IssueNumber: "1",
				Volume: Volume{
					ID:   67890,
					Name: "Batman",
				},
			},
			expectedError:     "",
			expectCacheCheck:  true,
			expectAPICall:     true,
			expectedCacheSize: 1,
		},
		{
			name:        "uses cached result when available",
			series:      "Batman",
			issueNumber: "1",
			mockResponses: []mockResponse{
				// First request populates cache
				{
					endpoint: "issues",
					params: map[string]string{
						"query": "Batman 1",
						"limit": "10",
						"sort":  "name:asc",
					},
					response: responseWithIssues([]Issue{
						{
							ID:          12345,
							Name:        "Batman #1",
							IssueNumber: "1",
							Volume: Volume{
								ID:   67890,
								Name: "Batman",
							},
						},
					}),
					statusCode: 200,
					err:        nil,
				},
				// Second request should use cache, not call API
			},
			expectedIssue: &Issue{
				ID:          12345,
				Name:        "Batman #1",
				IssueNumber: "1",
				Volume: Volume{
					ID:   67890,
					Name: "Batman",
				},
			},
			expectedError:     "",
			expectCacheCheck:  true,
			expectAPICall:     false, // No API call due to cache hit
			expectedCacheSize: 1,
		},
		{
			name:        "returns error when API fails",
			series:      "Batman",
			issueNumber: "1",
			mockResponses: []mockResponse{
				{
					endpoint: "issues",
					params: map[string]string{
						"query": "Batman 1",
						"limit": "10",
						"sort":  "name:asc",
					},
					response:   nil,
					statusCode: 500,
					err:        errors.New("API error"),
				},
			},
			expectedIssue:     nil,
			expectedError:     "API error",
			expectCacheCheck:  true,
			expectAPICall:     true,
			expectedCacheSize: 0, // No cache entry created when error occurs
		},
		{
			name:        "returns error when no results found",
			series:      "NonExistentSeries",
			issueNumber: "999",
			mockResponses: []mockResponse{
				{
					endpoint: "issues",
					params: map[string]string{
						"query": "NonExistentSeries 999",
						"limit": "10",
						"sort":  "name:asc",
					},
					response:   responseWithIssues([]Issue{}),
					statusCode: 200,
					err:        nil,
				},
			},
			expectedIssue:     nil,
			expectedError:     "no results found",
			expectCacheCheck:  true,
			expectAPICall:     true,
			expectedCacheSize: 0, // Empty results aren't cached
		},
		{
			name:        "finds best match when multiple results",
			series:      "Batman",
			issueNumber: "1",
			mockResponses: []mockResponse{
				{
					endpoint: "issues",
					params: map[string]string{
						"query": "Batman 1",
						"limit": "10",
						"sort":  "name:asc",
					},
					response: responseWithIssues([]Issue{
						{
							ID:          12345,
							Name:        "Some Other Comic",
							IssueNumber: "1",
							Volume: Volume{
								ID:   67890,
								Name: "Some Other Series",
							},
						},
						{
							ID:          54321,
							Name:        "Batman #1",
							IssueNumber: "1",
							Volume: Volume{
								ID:   98765,
								Name: "Batman",
							},
						},
					}),
					statusCode: 200,
					err:        nil,
				},
			},
			expectedIssue: &Issue{
				ID:          54321,
				Name:        "Batman #1",
				IssueNumber: "1",
				Volume: Volume{
					ID:   98765,
					Name: "Batman",
				},
			},
			expectedError:     "",
			expectCacheCheck:  true,
			expectAPICall:     true,
			expectedCacheSize: 1,
		},
		{
			name:        "case insensitive matching",
			series:      "batman", // lowercase
			issueNumber: "1",
			mockResponses: []mockResponse{
				{
					endpoint: "issues",
					params: map[string]string{
						"query": "batman 1",
						"limit": "10",
						"sort":  "name:asc",
					},
					response: responseWithIssues([]Issue{
						{
							ID:          12345,
							Name:        "Batman #1",
							IssueNumber: "1",
							Volume: Volume{
								ID:   67890,
								Name: "Batman", // Uppercase first letter
							},
						},
					}),
					statusCode: 200,
					err:        nil,
				},
			},
			expectedIssue: &Issue{
				ID:          12345,
				Name:        "Batman #1",
				IssueNumber: "1",
				Volume: Volume{
					ID:   67890,
					Name: "Batman",
				},
			},
			expectedError:     "",
			expectCacheCheck:  true,
			expectAPICall:     true,
			expectedCacheSize: 1,
		},
		{
			name:        "handles cache expiration",
			series:      "Batman",
			issueNumber: "1",
			mockResponses: []mockResponse{
				// Initial response
				{
					endpoint: "issues",
					params: map[string]string{
						"query": "Batman 1",
						"limit": "10",
						"sort":  "name:asc",
					},
					response: responseWithIssues([]Issue{
						{
							ID:          12345,
							Name:        "Batman #1",
							IssueNumber: "1",
							Volume: Volume{
								ID:   67890,
								Name: "Batman",
							},
						},
					}),
					statusCode: 200,
					err:        nil,
				},
				// Updated response after expiration
				{
					endpoint: "issues",
					params: map[string]string{
						"query": "Batman 1",
						"limit": "10",
						"sort":  "name:asc",
					},
					response: responseWithIssues([]Issue{
						{
							ID:          12345,
							Name:        "Batman #1 (Updated)",
							IssueNumber: "1",
							Volume: Volume{
								ID:   67890,
								Name: "Batman",
							},
						},
					}),
					statusCode: 200,
					err:        nil,
				},
			},
			expectedIssue: &Issue{
				ID:          12345,
				Name:        "Batman #1 (Updated)",
				IssueNumber: "1",
				Volume: Volume{
					ID:   67890,
					Name: "Batman",
				},
			},
			expectedError:     "",
			expectCacheCheck:  true,
			expectAPICall:     true, // API call made because cache expired
			expectedCacheSize: 1,    // Cache updated with new result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock client
			mockClient := new(MockClient)

			// Set up the mock client expectations
			for _, mockResp := range tt.mockResponses {
				mockClient.On("Get", mockResp.endpoint, mockResp.params).
					Return(mockResp.response, mockResp.statusCode, mockResp.err)
			}

			// Create service with mock client
			service := &Service{
				client:  mockClient,
				verbose: false,
				cache:   make(map[string]CacheEntry),
			}

			// If testing cache expiration, inject an expired cache entry
			if tt.name == "handles cache expiration" {
				cacheKey := getCacheKey(tt.series, tt.issueNumber)
				service.cache[cacheKey] = CacheEntry{
					Results: []Issue{
						{
							ID:          12345,
							Name:        "Batman #1",
							IssueNumber: "1",
							Volume: Volume{
								ID:   67890,
								Name: "Batman",
							},
						},
					},
					Timestamp: time.Now().Add(-25 * time.Hour), // Expired (24h TTL)
				}
			}

			// If testing cache hit, perform a first call to populate cache
			if tt.name == "uses cached result when available" {
				// First call to populate cache
				_, _ = service.GetIssue(tt.series, tt.issueNumber)

				// Reset mock to verify no additional calls are made
				mockClient.ExpectedCalls = nil
			}

			// Call the method being tested
			result, err := service.GetIssue(tt.series, tt.issueNumber)

			// Check error
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedIssue.ID, result.ID)
				assert.Equal(t, tt.expectedIssue.Name, result.Name)
				assert.Equal(t, tt.expectedIssue.IssueNumber, result.IssueNumber)
				assert.Equal(t, tt.expectedIssue.Volume.Name, result.Volume.Name)
			}

			// Check cache usage
			assert.Equal(t, tt.expectedCacheSize, len(service.cache))

			// Verify API call expectations
			mockClient.AssertExpectations(t)
		})
	}
}

func TestService_GetSeries(t *testing.T) {
	// Similar tests as GetIssue but for GetSeries which should handle the same business logic
	tests := []struct {
		name              string
		series            string
		issueNumber       string
		mockResponses     []mockResponse
		expectedIssue     *Issue
		expectedError     string
		expectCacheCheck  bool
		expectAPICall     bool
		expectedCacheSize int
	}{
		{
			name:        "retrieves series from API successfully",
			series:      "Batman",
			issueNumber: "1",
			mockResponses: []mockResponse{
				{
					endpoint: "issues",
					params: map[string]string{
						"query": "Batman 1",
						"limit": "10",
						"sort":  "name:asc",
					},
					response: responseWithIssues([]Issue{
						{
							ID:          12345,
							Name:        "Batman #1",
							IssueNumber: "1",
							Volume: Volume{
								ID:   67890,
								Name: "Batman",
							},
						},
					}),
					statusCode: 200,
					err:        nil,
				},
			},
			expectedIssue: &Issue{
				ID:          12345,
				Name:        "Batman #1",
				IssueNumber: "1",
				Volume: Volume{
					ID:   67890,
					Name: "Batman",
				},
			},
			expectedError:     "",
			expectCacheCheck:  true,
			expectAPICall:     true,
			expectedCacheSize: 1,
		},
		// Additional test cases can be similar to GetIssue
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock client
			mockClient := new(MockClient)

			// Set up the mock client expectations
			for _, mockResp := range tt.mockResponses {
				mockClient.On("Get", mockResp.endpoint, mockResp.params).
					Return(mockResp.response, mockResp.statusCode, mockResp.err)
			}

			// Create service with mock client
			service := &Service{
				client:  mockClient,
				verbose: false,
				cache:   make(map[string]CacheEntry),
			}

			// Call the method being tested
			result, err := service.GetSeries(tt.series, tt.issueNumber)

			// Check error
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedIssue.ID, result.ID)
				assert.Equal(t, tt.expectedIssue.Name, result.Name)
				assert.Equal(t, tt.expectedIssue.IssueNumber, result.IssueNumber)
				assert.Equal(t, tt.expectedIssue.Volume.Name, result.Volume.Name)
			}

			// Check cache usage
			assert.Equal(t, tt.expectedCacheSize, len(service.cache))

			// Verify API call expectations
			mockClient.AssertExpectations(t)
		})
	}
}

func TestService_FindBestMatch(t *testing.T) {
	tests := []struct {
		name         string
		series       string
		issueNumber  string
		issues       []Issue
		expectedID   int
		expectedName string
	}{
		{
			name:        "exact match on series and issue",
			series:      "Batman",
			issueNumber: "1",
			issues: []Issue{
				{
					ID:          12345,
					Name:        "Superman #5",
					IssueNumber: "5",
					Volume: Volume{
						Name: "Superman",
					},
				},
				{
					ID:          54321,
					Name:        "Batman #1",
					IssueNumber: "1",
					Volume: Volume{
						Name: "Batman",
					},
				},
			},
			expectedID:   54321,
			expectedName: "Batman #1",
		},
		{
			name:        "partial match on series",
			series:      "Batman",
			issueNumber: "1",
			issues: []Issue{
				{
					ID:          12345,
					Name:        "Superman #1",
					IssueNumber: "1",
					Volume: Volume{
						Name: "Superman",
					},
				},
				{
					ID:          54321,
					Name:        "The Batman Chronicles #1",
					IssueNumber: "1",
					Volume: Volume{
						Name: "The Batman Chronicles",
					},
				},
			},
			expectedID:   54321,
			expectedName: "The Batman Chronicles #1",
		},
		{
			name:        "exact match on issue number",
			series:      "Batman",
			issueNumber: "5",
			issues: []Issue{
				{
					ID:          12345,
					Name:        "Batman #4",
					IssueNumber: "4",
					Volume: Volume{
						Name: "Batman",
					},
				},
				{
					ID:          54321,
					Name:        "Batman #5",
					IssueNumber: "5",
					Volume: Volume{
						Name: "Batman",
					},
				},
			},
			expectedID:   54321,
			expectedName: "Batman #5",
		},
		{
			name:        "match with different case",
			series:      "batman",
			issueNumber: "1",
			issues: []Issue{
				{
					ID:          54321,
					Name:        "Batman #1",
					IssueNumber: "1",
					Volume: Volume{
						Name: "Batman",
					},
				},
			},
			expectedID:   54321,
			expectedName: "Batman #1",
		},
		{
			name:        "fallback to first result when no good match",
			series:      "Non-existent Series",
			issueNumber: "999",
			issues: []Issue{
				{
					ID:          12345,
					Name:        "Some Comic #1",
					IssueNumber: "1",
					Volume: Volume{
						Name: "Some Series",
					},
				},
			},
			expectedID:   12345,
			expectedName: "Some Comic #1",
		},
		{
			name:        "normalized issue number matching",
			series:      "Batman",
			issueNumber: "001",
			issues: []Issue{
				{
					ID:          54321,
					Name:        "Batman #1",
					IssueNumber: "1", // No leading zeros
					Volume: Volume{
						Name: "Batman",
					},
				},
			},
			expectedID:   54321,
			expectedName: "Batman #1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				verbose: false,
				cache:   make(map[string]CacheEntry),
			}

			// Normalize issue number as the service should
			normalizedIssueNumber := tt.issueNumber
			if len(normalizedIssueNumber) > 1 && normalizedIssueNumber[0] == '0' {
				for len(normalizedIssueNumber) > 1 && normalizedIssueNumber[0] == '0' {
					normalizedIssueNumber = normalizedIssueNumber[1:]
				}
			}

			result := service.findBestMatch(tt.issues, tt.series, normalizedIssueNumber)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedID, result.ID)
			assert.Equal(t, tt.expectedName, result.Name)
		})
	}
}

func TestService_NormalizeIssueNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single digit unchanged",
			input:    "1",
			expected: "1",
		},
		{
			name:     "double digit unchanged",
			input:    "42",
			expected: "42",
		},
		{
			name:     "removes single leading zero",
			input:    "01",
			expected: "1",
		},
		{
			name:     "removes multiple leading zeros",
			input:    "001",
			expected: "1",
		},
		{
			name:     "preserves zero",
			input:    "0",
			expected: "0",
		},
		{
			name:     "handles decimal points",
			input:    "1.5",
			expected: "1.5",
		},
		{
			name:     "removes leading zeros from decimal",
			input:    "01.5",
			expected: "1.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{}
			normalized := service.normalizeIssueNumber(tt.input)
			assert.Equal(t, tt.expected, normalized)
		})
	}
}

func TestService_GetCacheKey(t *testing.T) {
	tests := []struct {
		name        string
		series      string
		issueNumber string
		expected    string
	}{
		{
			name:        "basic key generation",
			series:      "Batman",
			issueNumber: "1",
			expected:    "batman:1",
		},
		{
			name:        "lowercase conversion",
			series:      "BATMAN",
			issueNumber: "1",
			expected:    "batman:1",
		},
		{
			name:        "issue normalization",
			series:      "Batman",
			issueNumber: "001",
			expected:    "batman:1",
		},
		{
			name:        "handles spaces",
			series:      "Batman Beyond",
			issueNumber: "1",
			expected:    "batman beyond:1",
		},
		{
			name:        "handles special characters",
			series:      "Batman & Robin",
			issueNumber: "1",
			expected:    "batman & robin:1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{}
			key := getCacheKey(tt.series, tt.issueNumber)
			assert.Equal(t, tt.expected, key)
		})
	}
}

// Helper types and functions for testing

type mockResponse struct {
	endpoint   string
	params     map[string]string
	response   []byte
	statusCode int
	err        error
}

// Temporary helper for test - will be implemented by the service
func getCacheKey(series, issueNumber string) string {
	normalizedIssue := issueNumber
	if len(normalizedIssue) > 1 && normalizedIssue[0] == '0' {
		for len(normalizedIssue) > 1 && normalizedIssue[0] == '0' {
			normalizedIssue = normalizedIssue[1:]
		}
	}
	return series + ":" + normalizedIssue
}

func responseWithIssues(issues []Issue) []byte {
	// Build mock response with specified issues
	response := map[string]interface{}{
		"status_code": 1,
		"results":     issues,
	}

	respBytes, _ := json.Marshal(response)
	return respBytes
}
