package parser

import (
	"context"
	"testing"

	"comic-parser/models"
)

func TestRegexParser_ReturnsLowConfidence(t *testing.T) {
	p := NewRegexParser()
	input := models.ParsedFilename{OriginalFilename: "test.cbz"}
	result, err := p.Parse(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Confidence == "high" {
		t.Errorf("expected not high confidence, got %s", result.Confidence)
	}
}

type mockParser struct {
	result models.ParsedFilename
	err    error
	called bool
}

func (m *mockParser) Parse(ctx context.Context, input models.ParsedFilename) (models.ParsedFilename, error) {
	m.called = true
	return m.result, m.err
}

func TestPipelineParser_PrimaryHighConfidence(t *testing.T) {
	primary := &mockParser{
		result: models.ParsedFilename{Confidence: "high", Title: "Primary"},
	}
	secondary := &mockParser{
		result: models.ParsedFilename{Confidence: "high", Title: "Secondary"},
	}

	pipeline := NewPipelineParser(primary, secondary)
	result, err := pipeline.Parse(context.Background(), models.ParsedFilename{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "Primary" {
		t.Errorf("expected primary result, got %s", result.Title)
	}
	if !primary.called {
		t.Error("primary parser should have been called")
	}
	if secondary.called {
		t.Error("secondary parser should not have been called")
	}
}

func TestPipelineParser_PrimaryLowConfidence(t *testing.T) {
	primary := &mockParser{
		result: models.ParsedFilename{Confidence: "low", Title: "Primary"},
	}
	secondary := &mockParser{
		result: models.ParsedFilename{Confidence: "high", Title: "Secondary"},
	}

	pipeline := NewPipelineParser(primary, secondary)
	result, err := pipeline.Parse(context.Background(), models.ParsedFilename{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "Secondary" {
		t.Errorf("expected secondary result, got %s", result.Title)
	}
	if !primary.called {
		t.Error("primary parser should have been called")
	}
	if !secondary.called {
		t.Error("secondary parser should have been called")
	}
}
