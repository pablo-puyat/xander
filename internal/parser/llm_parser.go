package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"comic-parser/internal/llm"
	"comic-parser/internal/models"
	"comic-parser/internal/prompts"
)

// LLMClient defines the interface for LLM interactions required by the parser.
type LLMClient interface {
	CompleteWithRetry(ctx context.Context, prompt string, maxRetries int, delay time.Duration) (string, error)
}

// LLMParser implements the Parser interface using an LLM.
type LLMParser struct {
	client            LLMClient
	retryAttempts     int
	retryDelaySeconds int
}

// NewLLMParser creates a new LLMParser.
func NewLLMParser(client LLMClient, retryAttempts int, retryDelaySeconds int) *LLMParser {
	return &LLMParser{
		client:            client,
		retryAttempts:     retryAttempts,
		retryDelaySeconds: retryDelaySeconds,
	}
}

// Parse implements the Parser interface.
// It uses an LLM to parse the filename.
func (p *LLMParser) Parse(ctx context.Context, input *models.ParsedFilename) (*models.ParsedFilename, error) {
	prompt := prompts.FilenameParsePrompt(input.OriginalFilename)

	response, err := p.client.CompleteWithRetry(
		ctx,
		prompt,
		p.retryAttempts,
		time.Duration(p.retryDelaySeconds)*time.Second,
	)
	if err != nil {
		return nil, fmt.Errorf("LLM completion: %w", err)
	}

	// Extract JSON from response
	jsonStr := llm.ExtractJSON(response)

	// Create a new struct to hold the result
	// We unmarshal into a new struct to ensure we get a fresh parse
	var parsed models.ParsedFilename
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w (response: %s)", err, response)
	}

	// Ensure OriginalFilename is preserved from the input
	parsed.OriginalFilename = input.OriginalFilename

	return &parsed, nil
}
