package selector

import (
	"context"
	"time"

	"comic-parser/internal/models"
)

// Selector defines the interface for selecting a match from ComicVine results.
type Selector interface {
	Select(ctx context.Context, parsed *models.ParsedFilename, candidates []models.ComicVineIssue) (*models.MatchResult, error)
}

// LLMClient defines the interface for LLM interactions needed by LLMSelector.
type LLMClient interface {
	CompleteWithRetry(ctx context.Context, prompt string, maxRetries int, delay time.Duration) (string, error)
}
