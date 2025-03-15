package parse

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ParseComicFilename extracts series, issue, year information from a filename
// This function accepts any string that follows comic naming patterns regardless of extension
func ParseComicFilename(filename string) (series string, issue string, year string, err error) {
	// Remove the extension (if any)
	baseFilename := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Common pattern: "Absolute Superman 003 (2025) (Digital) (Pyrate-DCP).cbz"
	//modernPattern := regexp.MustCompile(`^(.*?)\s+\((\d{4})\)\s(?:\s+\(.*?\))*$`)
	modernPattern := regexp.MustCompile(`(.+?)\(([^)]*)\)(?:\(([^)]*)\))*`)
	matches := modernPattern.FindStringSubmatch(baseFilename)
	if len(matches) >= 3 {
		return processModernPattern(matches)
		//series = strings.TrimSpace(matches[1])
		//year = matches[3]
		//issue = matches[2]
	}

	// If no pattern matches, return an error
	return "", "", "", fmt.Errorf("no match found for %s", filename)
}

func processModernPattern(matches []string) (series string, issue string, year string, err error) {
	series = strings.TrimSpace(matches[1])
	year = matches[3]
	issue = matches[2]

	if year == "" && len(issue) == 4 {
		year = issue
		issue = ""
	}

	if issue == "" {
		// If the issue is empty, the year is in the series
		pattern := regexp.MustCompile(`^.+?\s(\d{3})$`)
		seriesMatches := pattern.FindStringSubmatch(series)
		if len(seriesMatches) > 1 {
			issue = seriesMatches[1]
			series = strings.TrimSpace(strings.TrimSuffix(series, issue))
		}
	}
	return
	//return "", "", "", fmt.Errorf("unable to parse %s", matches[0])
}
