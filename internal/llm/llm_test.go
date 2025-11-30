package llm

import (
	"testing"
)

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "Markdown JSON",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "Text before and after",
			input:    "Here is the JSON:\n{\"key\": \"value\"}\nHope it helps.",
			expected: `{"key": "value"}`,
		},
		{
			name:     "Nested braces",
			input:    `{"key": {"nested": "value"}}`,
			expected: `{"key": {"nested": "value"}}`,
		},
		{
			name:     "Braces in strings",
			input:    `{"key": "value with { braces } inside"}`,
			expected: `{"key": "value with { braces } inside"}`,
		},
		{
			name:     "No JSON",
			input:    "Just text",
			expected: "Just text",
		},
		{
			name:     "Broken JSON start",
			input:    "some text { broken json",
			expected: "{ broken json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractJSON(tt.input)
			if got != tt.expected {
				t.Errorf("ExtractJSON() = %q, want %q", got, tt.expected)
			}
		})
	}
}
