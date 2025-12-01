package parser

import (
	"context"

	"comic-parser/internal/models"
)

// RegexParser implements the Parser interface using regular expressions.
// Currently it is a placeholder that passes the input through unchanged.
type RegexParser struct{}

// NewRegexParser creates a new RegexParser.
func NewRegexParser() *RegexParser {
	return &RegexParser{}
}

// Parse implements the Parser interface.
// It currently returns the input struct as-is, simulating a low-confidence match
// or a pass-through behavior.
func (p *RegexParser) Parse(ctx context.Context, input *models.ParsedFilename) (*models.ParsedFilename, error) {
	// In the future, this will use regex to extract info.
	// For now, it just returns the input.
	return input, nil
}
