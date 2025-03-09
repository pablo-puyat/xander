package parse

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
)

// Common errors
var (
	ErrNotComicFile     = errors.New("not a comic file")
	ErrInvalidFilename  = errors.New("invalid comic filename format")
)

// IsComicFile checks if the filename has a comic file extension
func IsComicFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".cbz" || ext == ".cbr"
}

// ParseComicFilename extracts series, issue, year, and publisher information from a filename
// This function accepts any filename regardless of extension
func ParseComicFilename(filename string) (series, issue, year, publisher string, err error) {
	// Remove the extension (if any)
	baseFilename := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Try to match different filename patterns
	
	// Pattern with publisher: "Publisher - Series (Year) #Issue.ext"
	publisherPattern := regexp.MustCompile(`^(.+)\s+-\s+(.+?)\s+\((\d{4})\)\s+#(\d+)$`)
	matches := publisherPattern.FindStringSubmatch(baseFilename)
	if len(matches) == 5 {
		publisher = matches[1]
		
		// Clean up series name if it contains "Vol. X"
		seriesName := matches[2]
		volumePattern := regexp.MustCompile(`(.*?)\s+Vol\.\s+\d+$`)
		volumeMatches := volumePattern.FindStringSubmatch(seriesName)
		if len(volumeMatches) == 2 {
			series = volumeMatches[1]
		} else {
			series = seriesName
		}
		
		year = matches[3]
		issue = matches[4]
		return
	}

	// Pattern without publisher: "Series (Year) #Issue.ext"
	standardPattern := regexp.MustCompile(`^(.*?)\s+\((\d{4})\)\s+#(\d+)$`)
	matches = standardPattern.FindStringSubmatch(baseFilename)
	if len(matches) == 4 {
		// Clean up series name if it contains "Vol. X"
		seriesName := matches[1]
		volumePattern := regexp.MustCompile(`(.*?)\s+Vol\.\s+\d+$`)
		volumeMatches := volumePattern.FindStringSubmatch(seriesName)
		if len(volumeMatches) == 2 {
			series = volumeMatches[1]
		} else {
			series = seriesName
		}
		
		year = matches[2]
		issue = matches[3]
		return
	}

	// If no pattern matches, return an error
	return "", "", "", "", ErrInvalidFilename
}