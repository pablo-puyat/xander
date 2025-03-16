package matchers

import (
	"testing"
	"xander/internal/comicvine"
)

// MockMatcher implements the Matcher interface for testing
type MockMatcher struct {
	name         string
	returnResult *comicvine.Result
}

func (m MockMatcher) Match(results []comicvine.Result, query string) *comicvine.Result {
	return m.returnResult
}

func (m MockMatcher) Name() string {
	return m.name
}

func TestMatchChain_FindBestMatch(t *testing.T) {
	testResults := []comicvine.Result{
		{Name: "Batman"},
		{Name: "Superman"},
		{Name: "Wonder Woman"},
	}

	// Test with empty results
	chain := NewMatchChain(false)
	result := chain.FindBestMatch([]comicvine.Result{}, "Superman")
	if result != nil {
		t.Errorf("Expected nil for empty results, got %v", result)
	}

	// Test with a match
	result = chain.FindBestMatch(testResults, "Superman")
	if result == nil {
		t.Error("Expected match for 'Superman', got nil")
	} else if result.Name != "Superman" {
		t.Errorf("Expected 'Superman', got %q", result.Name)
	}

	// Test with no match
	result = chain.FindBestMatch(testResults, "Flash")
	if result != nil {
		t.Errorf("Expected nil for non-matching query, got %v", result)
	}
}

func TestMatchChain_AddMatcher(t *testing.T) {
	chain := NewMatchChain(false)

	// Create a matcher that will always return a specific result
	expectedResult := &comicvine.Result{Name: "Custom Result"}
	customMatcher := MockMatcher{
		name:         "CustomMatcher",
		returnResult: expectedResult,
	}

	// Add the custom matcher
	chain.AddMatcher(customMatcher)

	// Test results where the default matchers won't find a match
	testResults := []comicvine.Result{
		{Name: "Batman"},
		{Name: "Superman"},
	}

	// The query "Flash" won't match any default matchers,
	// but our custom matcher will return a result anyway
	result := chain.FindBestMatch(testResults, "Flash")

	if result != expectedResult {
		t.Errorf("Expected result from custom matcher, got %v", result)
	}
}

func TestNewMatchChain(t *testing.T) {
	// Test that NewMatchChain creates a chain with the expected configuration
	chain := NewMatchChain(true)

	if chain == nil {
		t.Fatal("Expected non-nil MatchChain")
	}

	if !chain.verbose {
		t.Error("Expected verbose to be true")
	}

	// Check that there is at least one matcher (ExactMatcher)
	if len(chain.matchers) < 1 {
		t.Error("Expected at least one matcher in the chain")
	}
}
