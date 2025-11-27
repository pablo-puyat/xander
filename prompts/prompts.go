// Package prompts contains LLM prompt templates for comic parsing and matching.
// These prompts are critical to the application's accuracy and should be tuned carefully.
package prompts

import (
	"encoding/json"
	"fmt"

	"comic-parser/models"
)

// FilenameParsePrompt generates the prompt for parsing a comic filename.
// This prompt instructs the LLM to extract structured information from various filename formats.
func FilenameParsePrompt(filename string) string {
	return fmt.Sprintf(`You are a comic book filename parser. Your task is to extract structured information from comic book archive filenames (CBR/CBZ files).

Analyze the following filename and extract the comic title and issue number. Comic filenames come in many formats, such as:
- "Amazing Spider-Man 001 (2018).cbz"
- "Batman - The Long Halloween 01.cbr"  
- "X-Men v2 #45 (1995).cbz"
- "Saga 001 (2012) (Digital) (Zone-Empire).cbr"
- "The Walking Dead #100 (2012) (Digital).cbz"
- "Action_Comics_1000_(2018).cbr"
- "Invincible 001 (2003) (digital) (Son of Ultron-Empire).cbr"

Key patterns to recognize:
- Issue numbers may be preceded by #, No., or nothing
- Issue numbers may be zero-padded (001, 01, 1)
- Volume indicators: v1, v2, Vol. 1, Volume 2
- Years in parentheses: (2018), (1995)
- Publisher names sometimes appear
- Digital/scan group tags in parentheses at the end
- Underscores or hyphens used as word separators

FILENAME TO PARSE:
%s

Respond with ONLY a JSON object in this exact format (no markdown, no explanation):
{
  "title": "The main comic series title, cleaned up (e.g., 'Amazing Spider-Man', not 'Amazing_Spider-Man')",
  "issue_number": "The issue number as a simple string (e.g., '1', '100', '45.1')",
  "year": "Publication year if present, or empty string",
  "publisher": "Publisher if identifiable, or empty string",
  "volume_number": "Volume number if present (e.g., '2' for v2), or empty string",
  "confidence": "high/medium/low - your confidence in the extraction",
  "notes": "Any relevant notes about ambiguity or special cases"
}`, filename)
}

// ResultMatchPrompt generates the prompt for selecting the best ComicVine match.
// It presents the LLM with parsed information and search results to make an informed choice.
func ResultMatchPrompt(parsed models.ParsedFilename, results []models.ComicVineIssue) string {
	// Prepare a simplified view of the results for the LLM
	type SimpleResult struct {
		Index       int    `json:"index"`
		ID          int    `json:"id"`
		VolumeName  string `json:"volume_name"`
		IssueNumber string `json:"issue_number"`
		CoverDate   string `json:"cover_date"`
		Publisher   string `json:"publisher,omitempty"`
		URL         string `json:"url"`
	}

	simpleResults := make([]SimpleResult, len(results))
	for i, r := range results {
		simpleResults[i] = SimpleResult{
			Index:       i,
			ID:          r.ID,
			VolumeName:  r.Volume.Name,
			IssueNumber: r.IssueNumber,
			CoverDate:   r.CoverDate,
			Publisher:   r.Volume.Publisher,
			URL:         r.SiteDetailURL,
		}
	}

	resultsJSON, _ := json.MarshalIndent(simpleResults, "", "  ")

	return fmt.Sprintf(`You are a comic book matching expert. Your task is to select the best match from ComicVine search results for a given comic file.

ORIGINAL FILENAME: %s

PARSED INFORMATION:
- Title: %s
- Issue Number: %s
- Year: %s
- Publisher: %s
- Volume: %s
- Parser Notes: %s

COMICVINE SEARCH RESULTS:
%s

Your task:
1. Analyze each result against the parsed information
2. Select the BEST match based on:
   - Title/volume name similarity (most important)
   - Issue number match (must match exactly or very closely)
   - Year/cover date alignment (if available)
   - Publisher match (if known)
3. If no result is a good match, indicate that

Consider these matching rules:
- The volume name should match the comic title (accounting for variations like "The Amazing Spider-Man" vs "Amazing Spider-Man")
- Issue numbers must match (01 = 1 = 001)
- If a year is specified, the cover_date should be close (within 1-2 years to account for publication delays)
- Some comics have multiple volumes/series with the same name - prefer the one with matching year

Respond with ONLY a JSON object in this exact format (no markdown, no explanation):
{
  "selected_index": <index number of best match, or -1 if no good match>,
  "match_confidence": "high/medium/low/none",
  "reasoning": "Brief explanation of why this match was selected or why no match was found"
}`,
		parsed.OriginalFilename,
		parsed.Title,
		parsed.IssueNumber,
		parsed.Year,
		parsed.Publisher,
		parsed.VolumeNumber,
		parsed.Notes,
		string(resultsJSON))
}

// MatchResponse represents the LLM's response to the matching prompt.
type MatchResponse struct {
	SelectedIndex   int    `json:"selected_index"`
	MatchConfidence string `json:"match_confidence"`
	Reasoning       string `json:"reasoning"`
}
