package parser

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

// LLMClient defines the interface for LLM interactions needed by the parser.
type LLMClient interface {
	CompleteWithRetry(ctx context.Context, prompt string, maxRetries int, delay time.Duration) (string, error)
}

// LLMParser parses filenames using an LLM.
type LLMParser struct {
	client LLMClient
	cfg    *config.Config
}

// NewLLMParser creates a new LLMParser.
func NewLLMParser(client LLMClient, cfg *config.Config) *LLMParser {
	return &LLMParser{
		client: client,
		cfg:    cfg,
	}
}

// Parse implements the Parser interface.
func (p *LLMParser) Parse(ctx context.Context, input models.ParsedFilename) (models.ParsedFilename, error) {
	filename := input.OriginalFilename
	prompt := prompts.FilenameParsePrompt(filename)

	response, err := p.client.CompleteWithRetry(
		ctx,
		prompt,
		p.cfg.RetryAttempts,
		time.Duration(p.cfg.RetryDelaySeconds)*time.Second,
	)
	if err != nil {
		return models.ParsedFilename{}, fmt.Errorf("LLM completion: %w", err)
	}

	// Extract JSON from response
	jsonStr := llm.ExtractJSON(response)

	var parsed models.ParsedFilename
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return models.ParsedFilename{}, fmt.Errorf("parsing LLM response: %w (response: %s)", err, response)
	}

	// Ensure OriginalFilename is preserved from input
	parsed.OriginalFilename = filename
	return parsed, nil
}
