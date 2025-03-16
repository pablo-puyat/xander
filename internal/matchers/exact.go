package matchers

import (
	"strings"

	"xander/internal/comicvine"
)

// ExactMatcher implements exact name matching
type ExactMatcher struct{}

// Match returns a result if there's an exact name match
func (m ExactMatcher) Match(results []comicvine.Result, query string) *comicvine.Result {
	queryLower := strings.ToLower(query)
	for i, result := range results {
		if strings.ToLower(result.Name) == queryLower {
			return &results[i]
		}
	}
	return nil
}

// Name returns the name of this matcher
func (m ExactMatcher) Name() string {
	return "ExactMatcher"
}
