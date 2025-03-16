package matchers

import (
	"testing"

	"xander/internal/comicvine"
)

// This is a test stub for a future FuzzyMatcher implementation
// It's designed to test the test pattern for future matchers

// MockFuzzyMatcher is a test placeholder for a future matcher implementation
type MockFuzzyMatcher struct {
	MinimumScore float64
}

func (m MockFuzzyMatcher) Match(results []comicvine.Result, query string) *comicvine.Result {
	// This is a simple example of how a fuzzy matcher might work
	// It just returns the first result for test purposes
	// A real implementation would score matches by similarity
	if len(results) > 0 {
		return &results[0]
	}
	return nil
}

func (m MockFuzzyMatcher) Name() string {
	return "MockFuzzyMatcher"
}

func TestMockFuzzyMatcher(t *testing.T) {
	// Setup test data
	results := []comicvine.Result{
		{Name: "Batman"},
		{Name: "Superman"},
		{Name: "Wonder Woman"},
	}
	
	matcher := MockFuzzyMatcher{MinimumScore: 0.5}
	
	// Test basic match
	result := matcher.Match(results, "Btmn")
	if result == nil {
		t.Error("Expected match, got nil")
	} else if result.Name != "Batman" {
		t.Errorf("Expected 'Batman', got %q", result.Name)
	}
	
	// Test empty results
	result = matcher.Match([]comicvine.Result{}, "Batman")
	if result != nil {
		t.Errorf("Expected nil for empty results, got %v", result)
	}
	
	// Test name function
	name := matcher.Name()
	expected := "MockFuzzyMatcher"
	if name != expected {
		t.Errorf("Expected name %q, got %q", expected, name)
	}
}