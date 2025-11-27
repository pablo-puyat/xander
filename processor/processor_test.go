package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	"comic-parser/config"
	"comic-parser/models"
)

// MockLLMClient
type MockLLMClient struct {
	CompleteWithRetryFunc func(ctx context.Context, prompt string, maxRetries int, delay time.Duration) (string, error)
}

func (m *MockLLMClient) CompleteWithRetry(ctx context.Context, prompt string, maxRetries int, delay time.Duration) (string, error) {
	if m.CompleteWithRetryFunc != nil {
		return m.CompleteWithRetryFunc(ctx, prompt, maxRetries, delay)
	}
	return "", nil
}

func (m *MockLLMClient) Close() {}

// MockCVClient
type MockCVClient struct {
	SearchIssuesFunc func(ctx context.Context, title string, issueNumber string) ([]models.ComicVineIssue, error)
}

func (m *MockCVClient) SearchIssues(ctx context.Context, title string, issueNumber string) ([]models.ComicVineIssue, error) {
	if m.SearchIssuesFunc != nil {
		return m.SearchIssuesFunc(ctx, title, issueNumber)
	}
	return nil, nil
}

func (m *MockCVClient) Close() {}

func TestProcessFile(t *testing.T) {
	// Setup
	cfg := config.DefaultConfig()
	mockLLM := &MockLLMClient{}
	mockCV := &MockCVClient{}

	proc := &Processor{
		cfg:       cfg,
		llmClient: mockLLM,
		cvClient:  mockCV,
		verbose:   true,
	}

	// Mock LLM responses
	// First call is parsing, second call is matching
	callCount := 0
	mockLLM.CompleteWithRetryFunc = func(ctx context.Context, prompt string, maxRetries int, delay time.Duration) (string, error) {
		callCount++
		if callCount == 1 {
			// Parse response
			return `{"title": "Test Comic", "issue_number": "1", "confidence": "high"}`, nil
		}
		if callCount == 2 {
			// Match response
			return `{"selected_index": 0, "match_confidence": "high", "reasoning": "Perfect match"}`, nil
		}
		return "", errors.New("unexpected call")
	}

	// Mock CV response
	mockCV.SearchIssuesFunc = func(ctx context.Context, title string, issueNumber string) ([]models.ComicVineIssue, error) {
		return []models.ComicVineIssue{
			{
				ID:          123,
				Name:        "Test Comic",
				IssueNumber: "1",
				Volume:      models.VolumeRef{Name: "Test Comic"},
			},
		}, nil
	}

	// Run
	result, err := proc.ProcessFile(context.Background(), "Test Comic 001.cbz")

	// Assert
	if err != nil {
		t.Fatalf("ProcessFile returned error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("Result has error: %s", result.Error)
	}
	if !result.Success {
		t.Fatal("Result not success")
	}
	if result.Match == nil {
		t.Fatal("Match is nil")
	}
	if result.Match.ParsedInfo.Title != "Test Comic" {
		t.Errorf("Parsed Title = %s, want Test Comic", result.Match.ParsedInfo.Title)
	}
	if result.Match.SelectedIssue == nil {
		t.Fatal("SelectedIssue is nil")
	}
	if result.Match.SelectedIssue.ID != 123 {
		t.Errorf("SelectedIssue ID = %d, want 123", result.Match.SelectedIssue.ID)
	}
}

func TestProcessFile_ParseError(t *testing.T) {
	// Setup
	cfg := config.DefaultConfig()
	mockLLM := &MockLLMClient{}
	mockCV := &MockCVClient{}

	proc := &Processor{
		cfg:       cfg,
		llmClient: mockLLM,
		cvClient:  mockCV,
	}

	mockLLM.CompleteWithRetryFunc = func(ctx context.Context, prompt string, maxRetries int, delay time.Duration) (string, error) {
		return "", errors.New("llm error")
	}

	// Run
	result, _ := proc.ProcessFile(context.Background(), "file.cbz")

	if result.Error == "" {
		t.Fatal("Expected error in result, got none")
	}
}
