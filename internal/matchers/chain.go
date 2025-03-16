package matchers

import (
	"fmt"
	"xander/internal/comicvine"
)

// MatchChain manages a chain of matchers that will be tried in sequence
type MatchChain struct {
	matchers []Matcher
	verbose  bool
}

// NewMatchChain creates a new chain with default matchers
func NewMatchChain(verbose bool) *MatchChain {
	return &MatchChain{
		matchers: []Matcher{
			ExactMatcher{},
			//matchers.RelevanceMatcher{MinimumScore: 20},
			//matchers.FallbackMatcher{},
		},
		verbose: verbose,
	}
}

// AddMatcher adds a custom matcher to the chain
func (mc *MatchChain) AddMatcher(matcher Matcher) {
	mc.matchers = append(mc.matchers, matcher)
}

// FindBestMatch tries each matcher in sequence until one returns a result
func (mc *MatchChain) FindBestMatch(results []comicvine.Result, query string) *comicvine.Result {
	if len(results) == 0 {
		return nil
	}

	for _, matcher := range mc.matchers {
		if result := matcher.Match(results, query); result != nil {
			if mc.verbose {
				fmt.Printf("Match found using %s\n", matcher.Name())
			}
			return result
		}
	}

	return nil
}
