package parser

import (
	"context"
	"strings"

	"comic-parser/internal/models"
)

// ChainParser combines two parsers: a primary (usually Regex) and a fallback (usually LLM).
// If the primary parser fails to produce a high-confidence result, the fallback is used.
type ChainParser struct {
	primary  Parser
	fallback Parser
}

// NewChainParser creates a new ChainParser.
func NewChainParser(primary Parser, fallback Parser) *ChainParser {
	return &ChainParser{
		primary:  primary,
		fallback: fallback,
	}
}

// Parse implements the Parser interface.
func (p *ChainParser) Parse(ctx context.Context, input *models.ParsedFilename) (*models.ParsedFilename, error) {
	// Try primary parser
	result, err := p.primary.Parse(ctx, input)

	// Determine if we need to use the fallback parser
	shouldFallback := false
	if err != nil {
		// If primary errors, we fallback (assuming we want to try the next method)
		shouldFallback = true
	} else if result == nil {
		shouldFallback = true
	} else if strings.ToLower(result.Confidence) != "high" {
		shouldFallback = true
	}

	if shouldFallback {
		// Execute fallback parser
		// We pass the original input to ensure a clean attempt
		return p.fallback.Parse(ctx, input)
	}

	return result, nil
}
