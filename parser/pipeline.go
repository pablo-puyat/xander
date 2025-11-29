package parser

import (
	"context"

	"comic-parser/models"
)

// PipelineParser coordinates between multiple parsers (e.g., Regex -> LLM).
type PipelineParser struct {
	primary   Parser
	secondary Parser
}

// NewPipelineParser creates a new PipelineParser.
func NewPipelineParser(primary, secondary Parser) *PipelineParser {
	return &PipelineParser{
		primary:   primary,
		secondary: secondary,
	}
}

// Parse implements the Parser interface.
// It tries the primary parser first. If the result confidence is not "high",
// it falls back to the secondary parser.
func (p *PipelineParser) Parse(ctx context.Context, input models.ParsedFilename) (models.ParsedFilename, error) {
	result, err := p.primary.Parse(ctx, input)
	if err != nil {
		return models.ParsedFilename{}, err
	}

	if result.Confidence == "high" {
		return result, nil
	}

	return p.secondary.Parse(ctx, input)
}
