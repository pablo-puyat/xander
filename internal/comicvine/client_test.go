package comicvine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		verbose   bool
		wantNil   bool
		wantHTTP  bool
		wantError bool
	}{
		{
			name:     "creates client with valid API key",
			apiKey:   "test-key",
			verbose:  false,
			wantNil:  false,
			wantHTTP: true,
		},
		{
			name:     "creates verbose client",
			apiKey:   "test-key",
			verbose:  true,
			wantNil:  false,
			wantHTTP: true,
		},
		{
			name:     "fails with empty API key",
			apiKey:   "",
			verbose:  false,
			wantNil:  true,
			wantHTTP: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.apiKey, tt.verbose)

			if tt.wantNil && assert.Nil(t, client) {
				return
			}

			assert.NotNil(t, client)
			assert.Equal(t, tt.apiKey, client.apiKey)
			assert.Equal(t, tt.verbose, client.verbose)
			assert.NotNil(t, client.httpClient)
		})
	}
}

func TestClientGet(t *testing.T) {
	tests := []struct {
		name            string
		serverResponse  func(w http.ResponseWriter, r *http.Request)
		endpoint        string
		params          map[string]string
		wantErr         bool
		wantErrContains string
		wantStatusCode  int
		wantContains    string
	}{
		{
			name: "successful response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/issues", r.URL.Path)
				assert.Equal(t, "test-key", r.URL.Query().Get("api_key"))
				assert.Equal(t, "json", r.URL.Query().Get("format"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status_code":1,"results":[{"id":123}]}`))
			},
			endpoint: "issues",
			params: map[string]string{
				"query": "batman",
				"limit": "10",
			},
			wantErr:        false,
			wantStatusCode: 200,
			wantContains:   `"id":123`,
		},
		{
			name: "api error response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status_code":100,"error":"Invalid API key"}`))
			},
			endpoint: "issues",
			params: map[string]string{
				"query": "batman",
			},
			wantErr:         true,
			wantErrContains: "API error",
			wantStatusCode:  200,
		},
		{
			name: "http error response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`Server Error`))
			},
			endpoint: "issues",
			params: map[string]string{
				"query": "batman",
			},
			wantErr:         true,
			wantErrContains: "unexpected status code: 500",
			wantStatusCode:  500,
		},
		{
			name: "rate limit response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Header().Set("Retry-After", "60")
				w.Write([]byte(`Rate limited`))
			},
			endpoint: "issues",
			params: map[string]string{
				"query": "batman",
			},
			wantErr:         true,
			wantErrContains: "rate limit exceeded",
			wantStatusCode:  429,
		},
		{
			name: "invalid json response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`invalid json`))
			},
			endpoint: "issues",
			params: map[string]string{
				"query": "batman",
			},
			wantErr:         true,
			wantErrContains: "failed to parse",
			wantStatusCode:  200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client := &Client{
				apiKey:     "test-key",
				baseURL:    server.URL + "/api",
				httpClient: server.Client(),
				verbose:    true,
			}

			ctx := context.Background()
			response, statusCode, err := client.Get(ctx, tt.endpoint, tt.params)

			assert.Equal(t, tt.wantStatusCode, statusCode)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				responseStr := string(response)
				assert.Contains(t, responseStr, tt.wantContains)
			}
		})
	}
}

// TestRateLimiting tests the client's rate limiting functionality
func TestRateLimiting(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status_code":1,"results":[]}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:             "test-key",
		baseURL:            server.URL,
		httpClient:         server.Client(),
		verbose:            false,
		lastRequestTime:    time.Time{},
		minRequestInterval: 500 * time.Millisecond,
	}

	// Make several requests in a loop
	start := time.Now()
	for i := 0; i < 3; i++ {
		ctx := context.Background()
		_, _, err := client.Get(ctx, "test", nil)
		require.NoError(t, err)
	}
	duration := time.Since(start)

	// Verify time elapsed is at least the min interval * (requests-1)
	assert.GreaterOrEqual(t, duration, client.minRequestInterval*2,
		"Rate limiting should enforce minimum intervals between requests")
	assert.Equal(t, 3, requestCount, "All requests should have been processed")
}

// TestRequestCancellation tests that requests can be cancelled via context
func TestRequestCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow endpoint
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status_code":1}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: server.Client(),
		verbose:    false,
	}

	// Create a context with timeout shorter than the server delay
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, _, err := client.Get(ctx, "test", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestClientRequestFormatting tests that the client properly formats requests
func TestClientRequestFormatting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request URL components
		assert.Equal(t, "/test-endpoint", r.URL.Path)

		// Check query parameters
		query := r.URL.Query()
		assert.Equal(t, "test-key", query.Get("api_key"))
		assert.Equal(t, "json", query.Get("format"))
		assert.Equal(t, "Batman", query.Get("query"))
		assert.Equal(t, "10", query.Get("limit"))

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status_code":1,"results":[]}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: server.Client(),
		verbose:    false,
	}

	ctx := context.Background()
	params := map[string]string{
		"query": "Batman",
		"limit": "10",
	}

	_, _, err := client.Get(ctx, "test-endpoint", params)
	require.NoError(t, err)
}

// TestVerboseLogging tests the verbose logging option (indirectly)
func TestVerboseLogging(t *testing.T) {
	// This test doesn't really validate the log output,
	// just ensures the code path doesn't crash when verbose is on

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status_code":1,"results":[]}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: server.Client(),
		verbose:    true, // Enable verbose logging
	}

	ctx := context.Background()
	_, _, err := client.Get(ctx, "test", nil)
	require.NoError(t, err)

	// No assertion needed - if verbose logging code crashes, the test will fail
}
