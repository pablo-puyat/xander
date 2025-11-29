package parser

import (
	"context"

	"comic-parser/models"
)

// Parser defines the interface for parsing comic filenames.
// It takes a ParsedFilename struct (with OriginalFilename populated) and returns a populated ParsedFilename.
type Parser interface {
	Parse(ctx context.Context, input models.ParsedFilename) (models.ParsedFilename, error)
}
