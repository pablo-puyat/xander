package matchers_test

import (
	"testing"
	"xander/internal/comicvine"
	"xander/internal/matchers"
)

func TestExactMatcher_Match(t *testing.T) {
	tests := []struct {
		name     string
		results  []comicvine.Result
		query    string
		wantName string // The expected name of the matched result or empty if no match expected
	}{
		{
			name: "exact match",
			results: []comicvine.Result{
				{Name: "Batman"},
				{Name: "Superman"},
				{Name: "Wonder Woman"},
			},
			query:    "Superman",
			wantName: "Superman",
		},
		{
			name: "case insensitive match",
			results: []comicvine.Result{
				{Name: "Batman"},
				{Name: "Superman"},
				{Name: "Wonder Woman"},
			},
			query:    "superman",
			wantName: "Superman",
		},
		{
			name: "no match",
			results: []comicvine.Result{
				{Name: "Batman"},
				{Name: "Superman"},
				{Name: "Wonder Woman"},
			},
			query:    "Flash",
			wantName: "",
		},
		{
			name:     "empty results",
			results:  []comicvine.Result{},
			query:    "Batman",
			wantName: "",
		},
		{
			name: "match with extra data",
			results: []comicvine.Result{
				{
					Name:        "Batman",
					ComicVineID: 1,
					Series:      "Detective Comics",
				},
				{
					Name:        "Superman",
					ComicVineID: 2,
					Series:      "Action Comics",
				},
			},
			query:    "Superman",
			wantName: "Superman",
		},
	}

	matcher := ExactMatcher{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.Match(tt.results, tt.query)

			// If expected wantName is empty, the result should be nil
			if tt.wantName == "" {
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
				return
			}

			// If we get here, we expected a non-nil result
			if result == nil {
				t.Errorf("Expected a match with name %q, got nil", tt.wantName)
				return
			}

			if result.Name != tt.wantName {
				t.Errorf("Expected name %q, got %q", tt.wantName, result.Name)
			}

			// For the test with extra data, verify that the correct result was chosen
			if tt.name == "match with extra data" && tt.wantName == "Superman" {
				if result.ComicVineID != 2 {
					t.Errorf("Expected ComicVineID %d, got %d", 2, result.ComicVineID)
				}
				if result.Series != "Action Comics" {
					t.Errorf("Expected Series %q, got %q", "Action Comics", result.Series)
				}
			}
		})
	}
}

func TestExactMatcher_Name(t *testing.T) {
	matcher := ExactMatcher{}
	expected := "ExactMatcher"
	
	if name := matcher.Name(); name != expected {
		t.Errorf("Expected name %q, got %q", expected, name)
	}
}

func TestMatchChain(t *testing.T) {
	tests := []struct {
		name     string
		results  []comicvine.Result
		query    string
		wantName string // The expected name of the matched result or empty if no match expected
	}{
		{
			name: "finds exact match using chain",
			results: []comicvine.Result{
				{Name: "Batman"},
				{Name: "Superman"},
				{Name: "Wonder Woman"},
			},
			query:    "Superman",
			wantName: "Superman",
		},
		{
			name:     "empty results returns nil",
			results:  []comicvine.Result{},
			query:    "Batman",
			wantName: "",
		},
		{
			name: "no match returns nil",
			results: []comicvine.Result{
				{Name: "Batman"},
				{Name: "Superman"},
			},
			query:    "Flash",
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := matchers.NewMatchChain(false)
			result := chain.FindBestMatch(tt.results, tt.query)

			// If expected wantName is empty, the result should be nil
			if tt.wantName == "" {
				if result != nil {
					t.Errorf("Expected nil result, got %v", result)
				}
				return
			}

			// If we get here, we expected a non-nil result
			if result == nil {
				t.Errorf("Expected a match with name %q, got nil", tt.wantName)
				return
			}

			if result.Name != tt.wantName {
				t.Errorf("Expected name %q, got %q", tt.wantName, result.Name)
			}
		})
	}
}

func TestMatchChain_AddMatcher(t *testing.T) {
	chain := matchers.NewMatchChain(false)
	
	// Create a custom matcher for testing
	type customMatcher struct{}
	
	called := false
	
	// Implement the Matcher interface
	func (m customMatcher) Match(results []comicvine.Result, query string) *comicvine.Result {
		called = true
		return nil
	}
	
	func (m customMatcher) Name() string {
		return "CustomMatcher"
	}
	
	// Add the custom matcher to the chain
	chain.AddMatcher(customMatcher{})
	
	// Set up test data
	testResults := []comicvine.Result{
		{Name: "Batman"},
	}
	
	// Run the chain
	result := chain.FindBestMatch(testResults, "Flash")
	
	// Since we return nil from our custom matcher, the result should be nil
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
	
	// Verify our custom matcher was called
	if !called {
		t.Error("Custom matcher was not called")
	}
}