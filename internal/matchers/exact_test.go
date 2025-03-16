package matchers

import (
	"testing"

	"xander/internal/comicvine"
)

func TestExactMatcher_Match(t *testing.T) {
	// Setup test data
	results := []comicvine.Result{
		{Name: "Batman"},
		{Name: "Superman"},
		{Name: "Wonder Woman"},
	}
	
	matcher := ExactMatcher{}
	
	// Test exact match
	result := matcher.Match(results, "Superman")
	if result == nil {
		t.Error("Expected match for 'Superman', got nil")
	} else if result.Name != "Superman" {
		t.Errorf("Expected 'Superman', got %q", result.Name)
	}
	
	// Test case insensitive match
	result = matcher.Match(results, "superman")
	if result == nil {
		t.Error("Expected case-insensitive match for 'superman', got nil")
	} else if result.Name != "Superman" {
		t.Errorf("Expected 'Superman', got %q", result.Name)
	}
	
	// Test no match
	result = matcher.Match(results, "Flash")
	if result != nil {
		t.Errorf("Expected nil for non-matching query, got %v", result)
	}
	
	// Test empty results
	result = matcher.Match([]comicvine.Result{}, "Batman")
	if result != nil {
		t.Errorf("Expected nil for empty results, got %v", result)
	}
}

func TestExactMatcher_Name(t *testing.T) {
	matcher := ExactMatcher{}
	expected := "ExactMatcher"
	
	name := matcher.Name()
	if name != expected {
		t.Errorf("Expected name %q, got %q", expected, name)
	}
}