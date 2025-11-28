package parser

import (
	"testing"
)

func TestRegexParser_Parse(t *testing.T) {
	tests := []struct {
		filename string
		wantTitle string
		wantIssue string
		wantYear  string
		wantNotesContain []string // Just check if notes contain specific substrings
	}{
		{
			filename:  "Aquaman 009 (2025) (Digital) (Pyrate-DCP).cbz",
			wantTitle: "Aquaman",
			wantIssue: "009",
			wantYear:  "2025",
			wantNotesContain: []string{"(Digital)", "(Pyrate-DCP)"},
		},
		{
			filename:  "2000AD prog 2429 (2025) (digital) (Minutemen-juvecube).cbr",
			wantTitle: "2000AD prog",
			wantIssue: "2429",
			wantYear:  "2025",
			wantNotesContain: []string{"(digital)", "(Minutemen-juvecube)"},
		},
		{
			filename:  "2000AD 2418 (2025) (Digital-Empire).cbr",
			wantTitle: "2000AD",
			wantIssue: "2418",
			wantYear:  "2025",
			wantNotesContain: []string{"(Digital-Empire)"},
		},
		{
			filename:  "Wesley Dodds - The Sandman (2024) (digital) (Son of Ultron-Empire).cbr",
			wantTitle: "Wesley Dodds - The Sandman",
			wantIssue: "",
			wantYear:  "2024",
			wantNotesContain: []string{"(digital)", "(Son of Ultron-Empire)"},
		},
		{
			filename:  "World of Archie Double Digest 143.cbz",
			wantTitle: "World of Archie Double Digest",
			wantIssue: "143",
			wantYear:  "",
			wantNotesContain: []string{},
		},
		{
			filename:  "Arcana Royale 02 (of 04) (2025) (digital) (Son of Ultron-Empire).cbr",
			wantTitle: "Arcana Royale",
			wantIssue: "02",
			wantYear:  "2025",
			wantNotesContain: []string{"(of 04)", "(digital)", "(Son of Ultron-Empire)"},
		},
		{
			filename: "Title #100.cbz",
			wantTitle: "Title",
			wantIssue: "100",
			wantYear: "",
			wantNotesContain: nil,
		},
		{
			filename: "Saga 001 (Digital).cbz",
			wantTitle: "Saga",
			wantIssue: "001",
			wantYear: "", // No year
			wantNotesContain: []string{"(Digital)"},
		},
	}

	p := NewRegexParser()

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got, err := p.Parse(tt.filename)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			if got.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, tt.wantTitle)
			}
			if got.IssueNumber != tt.wantIssue {
				t.Errorf("IssueNumber = %q, want %q", got.IssueNumber, tt.wantIssue)
			}
			if got.Year != tt.wantYear {
				t.Errorf("Year = %q, want %q", got.Year, tt.wantYear)
			}

			for _, note := range tt.wantNotesContain {
				// Simple check if note substring exists in got.Notes
				// Notes are joined by space.
				found := false
				if note == "" { continue }
				// We expect exact match of the group string
				// Because implementation stores "(Digital)"
				// But we join with space.
				// Check strict containment
				if contains(got.Notes, note) {
					found = true
				}

				if !found {
					t.Errorf("Notes = %q, want to contain %q", got.Notes, note)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		containsSubstring(s, " " + substr + " ")))) ||
		containsSubstring(s, substr) // Simpler: just strings.Contains
}

func containsSubstring(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
