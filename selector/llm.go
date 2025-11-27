package selector

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"comic-parser/config"
	"comic-parser/llm"
	"comic-parser/models"
	"comic-parser/prompts"
)

// LLMSelector uses an LLM to select the best match from candidates.
type LLMSelector struct {
	client LLMClient
	cfg    *config.Config
}

// NewLLMSelector creates a new LLMSelector.
func NewLLMSelector(client LLMClient, cfg *config.Config) *LLMSelector {
	return &LLMSelector{
		client: client,
		cfg:    cfg,
	}
}

// Select implements the Selector interface.
func (s *LLMSelector) Select(ctx context.Context, parsed *models.ParsedFilename, issues []models.ComicVineIssue) (*models.MatchResult, error) {
	result := &models.MatchResult{
		OriginalFilename: parsed.OriginalFilename,
		ParsedInfo:       *parsed,
	}

	if len(issues) == 0 {
		result.MatchConfidence = "none"
		result.Reasoning = "No results found in ComicVine"
		return result, nil
	}

	prompt := prompts.ResultMatchPrompt(*parsed, issues)

	response, err := s.client.CompleteWithRetry(
		ctx,
		prompt,
		s.cfg.RetryAttempts,
		time.Duration(s.cfg.RetryDelaySeconds)*time.Second,
	)
	if err != nil {
		return nil, fmt.Errorf("LLM completion: %w", err)
	}

	// Extract JSON from response
	jsonStr := llm.ExtractJSON(response)

	var matchResp prompts.MatchResponse
	if err := json.Unmarshal([]byte(jsonStr), &matchResp); err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w (response: %s)", err, response)
	}

	result.MatchConfidence = matchResp.MatchConfidence
	result.Reasoning = matchResp.Reasoning

	if matchResp.SelectedIndex >= 0 && matchResp.SelectedIndex < len(issues) {
		selectedIssue := issues[matchResp.SelectedIndex]
		result.SelectedIssue = &selectedIssue
		result.ComicVineID = selectedIssue.ID
		result.ComicVineURL = selectedIssue.SiteDetailURL
	}

	return result, nil
}
