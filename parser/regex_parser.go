package parser

import (
	"context"

	"comic-parser/models"
)

// RegexParser parses filenames using regex patterns.
// Currently a stub that triggers fallback to LLM.
type RegexParser struct{}

// NewRegexParser creates a new RegexParser.
func NewRegexParser() *RegexParser {
	return &RegexParser{}
}

// Parse implements the Parser interface.
func (p *RegexParser) Parse(ctx context.Context, input models.ParsedFilename) (models.ParsedFilename, error) {
	// Stub: Return input with low confidence to trigger fallback to LLM
	output := input
	output.Confidence = "low"
	return output, nil
}
