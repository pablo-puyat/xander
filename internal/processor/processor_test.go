package processor

import (
	"context"
	"testing"
	"time"

	"comic-parser/internal/config"
	"comic-parser/internal/models"
)

// MockLLMClient implements LLMClient
type MockLLMClient struct {
	CompleteFunc func(ctx context.Context, prompt string) (string, error)
}

func (m *MockLLMClient) CompleteWithRetry(ctx context.Context, prompt string, maxRetries int, delay time.Duration) (string, error) {
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, prompt)
	}
	return "", nil
}

func (m *MockLLMClient) Close() {}

// MockCVClient implements CVClient
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

// MockSelector implements selector.Selector
type MockSelector struct {
	SelectFunc func(ctx context.Context, parsed *models.ParsedFilename, candidates []models.ComicVineIssue) (*models.MatchResult, error)
}

func (m *MockSelector) Select(ctx context.Context, parsed *models.ParsedFilename, candidates []models.ComicVineIssue) (*models.MatchResult, error) {
	if m.SelectFunc != nil {
		return m.SelectFunc(ctx, parsed, candidates)
	}
	return nil, nil
}

func TestProcessor_ProcessFile(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		mockLLMResp   string
		mockIssues    []models.ComicVineIssue
		mockMatch     *models.MatchResult
		expectedError bool
		checkMatch    func(*testing.T, *models.ProcessingResult)
	}{
		{
			name:     "Successful match",
			filename: "Amazing Spider-Man 001.cbz",
			mockLLMResp: `{
				"title": "Amazing Spider-Man",
				"issue_number": "1",
				"year": "2018",
				"confidence": "high"
			}`,
			mockIssues: []models.ComicVineIssue{
				{ID: 1, Name: "Amazing Spider-Man", IssueNumber: "1"},
			},
			mockMatch: &models.MatchResult{
				MatchConfidence: "high",
				SelectedIssue:   &models.ComicVineIssue{ID: 1, Name: "Amazing Spider-Man"},
			},
			expectedError: false,
			checkMatch: func(t *testing.T, res *models.ProcessingResult) {
				if !res.Success {
					t.Error("Expected success")
				}
				if res.Match == nil {
					t.Fatal("Expected match result")
				}
				if res.Match.MatchConfidence != "high" {
					t.Errorf("Expected high confidence, got %s", res.Match.MatchConfidence)
				}
			},
		},
		{
			name:     "LLM Parse Error",
			filename: "Broken.cbz",
			mockLLMResp: `invalid json`,
			expectedError: false, // ProcessFile captures error in result
			checkMatch: func(t *testing.T, res *models.ProcessingResult) {
				if res.Success {
					t.Error("Expected failure")
				}
				if res.Error == "" {
					t.Error("Expected error message")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()

			llmClient := &MockLLMClient{
				CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
					return tt.mockLLMResp, nil
				},
			}

			cvClient := &MockCVClient{
				SearchIssuesFunc: func(ctx context.Context, title string, issueNumber string) ([]models.ComicVineIssue, error) {
					return tt.mockIssues, nil
				},
			}

			sel := &MockSelector{
				SelectFunc: func(ctx context.Context, parsed *models.ParsedFilename, candidates []models.ComicVineIssue) (*models.MatchResult, error) {
					// Need to populate ParsedInfo in result for consistency
					res := tt.mockMatch
					if res != nil {
						res.ParsedInfo = *parsed
					}
					return res, nil
				},
			}

			proc := NewProcessor(cfg, llmClient, cvClient, sel)
			ctx := context.Background()

			result, err := proc.ProcessFile(ctx, tt.filename)

			if (err != nil) != tt.expectedError {
				t.Errorf("ProcessFile() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if tt.checkMatch != nil {
				tt.checkMatch(t, result)
			}
		})
	}
}
