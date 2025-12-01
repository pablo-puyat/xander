// Package parser provides interfaces and implementations for parsing comic filenames.
package parser

import (
	"context"

	"comic-parser/internal/models"
)

// Parser defines the interface for parsing comic filenames.
// It takes a ParsedFilename struct (containing the original filename) and returns
// a parsed version of it (or a new struct with populated fields).
type Parser interface {
	Parse(ctx context.Context, input *models.ParsedFilename) (*models.ParsedFilename, error)
}
