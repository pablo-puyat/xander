// Package llm provides an Anthropic API client for LLM interactions.
// It handles communication with Claude models for parsing and matching operations.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"comic-parser/config"
)

const (
	// API configuration
	anthropicVersion = "2023-06-01"
	contentTypeJSON  = "application/json"
	headerAPIKey     = "x-api-key"
	headerVersion    = "anthropic-version"

	// HTTP client settings
	defaultHTTPTimeout = 60 * time.Second
)

// Client is an Anthropic API client.
type Client struct {
	apiKey      string
	baseURL     string
	model       string
	maxTokens   int
	httpClient  *http.Client
	rateLimiter *time.Ticker
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request represents an Anthropic API request
type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

// ContentBlock represents a content block in the response
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Response represents an Anthropic API response
type Response struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence"`
	Usage        Usage          `json:"usage"`
}

// Usage represents token usage in the response
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ErrorResponse represents an error from the Anthropic API
type ErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient creates a new Anthropic API client.
func NewClient(cfg *config.Config) *Client {
	// Calculate rate limit interval
	limit := cfg.RateLimitPerMin
	if limit <= 0 {
		limit = 30 // Safe default
	}
	interval := time.Minute / time.Duration(limit)

	return &Client{
		apiKey:    cfg.AnthropicAPIKey,
		baseURL:   cfg.AnthropicAPIBaseURL,
		model:     cfg.AnthropicModel,
		maxTokens: cfg.AnthropicMaxTokens,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		rateLimiter: time.NewTicker(interval),
	}
}

// Close cleans up client resources.
func (c *Client) Close() {
	if c.rateLimiter != nil {
		c.rateLimiter.Stop()
	}
}

// Complete sends a completion request to the Anthropic API
func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	// Respect rate limit
	if c.rateLimiter != nil {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-c.rateLimiter.C:
			// Proceed
		}
	}

	req := Request{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	return c.doRequest(ctx, req)
}

// CompleteWithRetry sends a completion request with retry logic
func (c *Client) CompleteWithRetry(ctx context.Context, prompt string, maxRetries int, delay time.Duration) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: delay * 2^(attempt-1)
			backoff := delay * time.Duration(math.Pow(2, float64(attempt-1)))

			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		result, err := c.Complete(ctx, prompt)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if strings.Contains(err.Error(), "invalid_api_key") ||
			strings.Contains(err.Error(), "authentication") {
			return "", err
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

func (c *Client) doRequest(ctx context.Context, req Request) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", contentTypeJSON)
	httpReq.Header.Set(headerAPIKey, c.apiKey)
	httpReq.Header.Set(headerVersion, anthropicVersion)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return "", fmt.Errorf("API error (status %d): %s - %s",
				resp.StatusCode, errResp.Error.Type, errResp.Error.Message)
		}
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("empty response content")
	}

	// Concatenate all text blocks
	var result strings.Builder
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}

	return result.String(), nil
}

// ExtractJSON extracts JSON from LLM response that might have extra text.
// It handles markdown code blocks and finds valid JSON object boundaries using json.Decoder.
func ExtractJSON(response string) string {
	// Try to find JSON object boundaries
	response = strings.TrimSpace(response)

	// Remove markdown code blocks if present
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		if idx := strings.LastIndex(response, "```"); idx != -1 {
			response = response[:idx]
		}
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		if idx := strings.LastIndex(response, "```"); idx != -1 {
			response = response[:idx]
		}
	}

	response = strings.TrimSpace(response)

	// Find the JSON object start
	start := strings.Index(response, "{")
	if start == -1 {
		return response
	}

	// Use json.Decoder to find the end of the object intelligently
	decoder := json.NewDecoder(strings.NewReader(response[start:]))
	var v json.RawMessage
	if err := decoder.Decode(&v); err != nil {
		// If decoding fails, fall back to returning everything from start
		return response[start:]
	}

	return string(v)
}
