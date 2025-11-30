package prompts

import (
	"strings"
	"testing"

	"comic-parser/internal/models"
)

func TestFilenameParsePrompt(t *testing.T) {
	filename := "Test Comic 001.cbz"
	prompt := FilenameParsePrompt(filename)

	if !strings.Contains(prompt, filename) {
		t.Errorf("FilenameParsePrompt() does not contain filename %q", filename)
	}
	if !strings.Contains(prompt, "FILENAME TO PARSE:") {
		t.Error("FilenameParsePrompt() missing 'FILENAME TO PARSE:' label")
	}
	if !strings.Contains(prompt, "Respond with ONLY a JSON object") {
		t.Error("FilenameParsePrompt() missing JSON instruction")
	}
}

func TestResultMatchPrompt(t *testing.T) {
	parsed := models.ParsedFilename{
		OriginalFilename: "Test Comic 001.cbz",
		Title:            "Test Comic",
		IssueNumber:      "1",
		Year:             "2023",
		Publisher:        "TestPub",
		VolumeNumber:     "v1",
		Notes:            "Some notes",
	}

	results := []models.ComicVineIssue{
		{
			ID:          123,
			Name:        "Test Comic",
			IssueNumber: "1",
			Volume: models.VolumeRef{
				Name: "Test Comic",
			},
		},
	}

	prompt := ResultMatchPrompt(parsed, results)

	// Check parsed info presence
	if !strings.Contains(prompt, parsed.OriginalFilename) {
		t.Error("ResultMatchPrompt() missing OriginalFilename")
	}
	if !strings.Contains(prompt, parsed.Title) {
		t.Error("ResultMatchPrompt() missing Title")
	}
	if !strings.Contains(prompt, parsed.IssueNumber) {
		t.Error("ResultMatchPrompt() missing IssueNumber")
	}
	if !strings.Contains(prompt, parsed.Year) {
		t.Error("ResultMatchPrompt() missing Year")
	}
	if !strings.Contains(prompt, parsed.Publisher) {
		t.Error("ResultMatchPrompt() missing Publisher")
	}

	// Check results presence (partial check via ID)
	if !strings.Contains(prompt, "123") {
		t.Error("ResultMatchPrompt() missing result ID")
	}
}
