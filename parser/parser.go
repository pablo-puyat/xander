package parser

import (
	"comic-parser/models"
	"path/filepath"
	"regexp"
	"strings"
)

// Parser defines the interface for filename parsers.
type Parser interface {
	Parse(filename string) (*models.ParsedFilename, error)
}

// RegexParser implements a regex-based filename parser.
type RegexParser struct{}

// NewRegexParser creates a new RegexParser.
func NewRegexParser() *RegexParser {
	return &RegexParser{}
}

// Compiled regexes for performance
var (
	// Matches trailing parenthesized groups: (Group1) (Group2)... at the end of string
	reTrailingGroups = regexp.MustCompile(`(?:\s*\([^)]+\))+$`)
	// Matches a single parenthesized group content including parens
	reGroup = regexp.MustCompile(`\(([^)]+)\)`)
	// Matches a year: (2023) or 2023
	reYear = regexp.MustCompile(`^\(?(\d{4})\)?$`)
	// Matches issue number pattern: #123
	reIssueHash = regexp.MustCompile(`#\s*(\S+)`)
	// Matches number at the end of the string
	reEndNumber = regexp.MustCompile(`\s+(\d+)$`)
)

// Parse implements the Parser interface.
func (p *RegexParser) Parse(filename string) (*models.ParsedFilename, error) {
	parsed := &models.ParsedFilename{
		OriginalFilename: filename,
		Confidence:       "high", // Default to high if we extract something meaningful
	}

	// 1. Strip extension
	ext := filepath.Ext(filename)
	cleanName := strings.TrimSuffix(filename, ext)

	// 2. Extract trailing parenthesized groups
	var notes []string

	// Find the chunk of text that represents all trailing groups
	trailingLoc := reTrailingGroups.FindStringIndex(cleanName)
	var titlePart string

	if trailingLoc != nil {
		titlePart = strings.TrimSpace(cleanName[:trailingLoc[0]])
		groupsText := cleanName[trailingLoc[0]:]

		// Parse individual groups
		// We use FindAllString to get "(Group1)", "(Group2)"
		matches := reGroup.FindAllString(groupsText, -1)
		for _, match := range matches {
			content := strings.Trim(match, "()")

			// Check if it's a Year (and we haven't found one yet)
			if parsed.Year == "" && reYear.MatchString(match) {
				parsed.Year = content
				continue // Don't add year to notes
			}

			// Check if it's "of XX" (Total Issues) -> Notes
			// The user said "put in notes". We add it to notes list.

			// Add to notes
			notes = append(notes, match)
		}
	} else {
		titlePart = strings.TrimSpace(cleanName)
	}

	// 3. Parse Title and Issue from titlePart
	// Strategy:
	// a. If contains '#', split.
	// b. If ends in number, split.
	// c. Else, everything is Title (Graphic Novel case).

	// Check for #
	if loc := reIssueHash.FindStringIndex(titlePart); loc != nil {
		// Found #Issue
		// titlePart is "Title #Issue stuff" (but we handled trailing stuff already,
		// however "Title #123" might be the whole titlePart)
		// Actually reIssueHash matches `#\s*(\S+)`.
		// We assume the Issue number is the match.
		// But wait, "Title #123" -> Title="Title", Issue="123".

		// Let's use Split.
		parts := strings.SplitN(titlePart, "#", 2)
		parsed.Title = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			// Take the first word after # as issue
			issuePart := strings.TrimSpace(parts[1])
			// Issue number should be a simple string.
			// If issuePart is "001", that's it.
			parsed.IssueNumber = issuePart
		}
	} else {
		// Check for number at end
		// reEndNumber matches `\s+(\d+)$`
		match := reEndNumber.FindStringSubmatch(titlePart)
		if match != nil {
			// match[0] is " 123", match[1] is "123"
			parsed.IssueNumber = match[1]
			parsed.Title = strings.TrimSpace(titlePart[:len(titlePart)-len(match[0])])

			// Special handling for "2000AD prog 2429"
			// Title "2000AD prog", Issue "2429".
			// Our logic yields: Title="2000AD prog", Issue="2429". Correct.
		} else {
			// No number at end -> Graphic Novel or implicit
			parsed.Title = titlePart
			parsed.IssueNumber = ""
		}
	}

	if len(notes) > 0 {
		parsed.Notes = strings.Join(notes, " ")
	}

	// Fallback logic check:
	// If Title is empty, this parser failed significantly.
	if parsed.Title == "" {
		// If we couldn't even find a title (e.g. filename was just "(2025).cbz"),
		// we probably shouldn't return a "success" result.
		// However, the caller will decide.
		// But typically empty title is invalid.
		parsed.Confidence = "low"
	}

	return parsed, nil
}
