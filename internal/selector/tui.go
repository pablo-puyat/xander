package selector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"comic-parser/internal/models"
)

// TUISelector presents candidates to the user via CLI.
type TUISelector struct {
	mu sync.Mutex
}

// NewTUISelector creates a new TUISelector.
func NewTUISelector() *TUISelector {
	return &TUISelector{}
}

// Select implements the Selector interface.
func (s *TUISelector) Select(ctx context.Context, parsed *models.ParsedFilename, issues []models.ComicVineIssue) (*models.MatchResult, error) {
	// Lock to ensure only one interaction happens at a time
	s.mu.Lock()
	defer s.mu.Unlock()

	result := &models.MatchResult{
		OriginalFilename: parsed.OriginalFilename,
		ParsedInfo:       *parsed,
	}

	if len(issues) == 0 {
		fmt.Printf("\n--- No Results Found ---\n")
		fmt.Printf("File: %s\n", parsed.OriginalFilename)
		fmt.Printf("Parsed: %s #%s (%s)\n", parsed.Title, parsed.IssueNumber, parsed.Year)
		fmt.Println("No candidates returned from ComicVine.")
		fmt.Println("Press Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')

		result.MatchConfidence = "none"
		result.Reasoning = "No results found in ComicVine"
		return result, nil
	}

	fmt.Printf("\n==================================================\n")
	fmt.Printf("File: %s\n", parsed.OriginalFilename)
	fmt.Printf("Parsed: Title='%s' Issue='%s' Year='%s' Publisher='%s'\n",
		parsed.Title, parsed.IssueNumber, parsed.Year, parsed.Publisher)
	fmt.Printf("--------------------------------------------------\n")
	fmt.Printf("Select a match (0 for No Match):\n")

	for i, issue := range issues {
		fmt.Printf("[%d] %s #%s (%s) - %s\n",
			i+1, issue.Volume.Name, issue.IssueNumber, issue.CoverDate, issue.Volume.Publisher)
	}
	fmt.Printf("--------------------------------------------------\n")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter selection [0-", len(issues), "]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		val, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}

		if val == 0 {
			result.MatchConfidence = "none"
			result.Reasoning = "User selected No Match"
			fmt.Println("Marked as No Match.")
			return result, nil
		}

		if val > 0 && val <= len(issues) {
			selectedIssue := issues[val-1]
			result.SelectedIssue = &selectedIssue
			result.ComicVineID = selectedIssue.ID
			result.ComicVineURL = selectedIssue.SiteDetailURL
			result.MatchConfidence = "high" // User manually selected it
			result.Reasoning = "User manual selection"
			fmt.Printf("Selected: %s #%s\n", selectedIssue.Volume.Name, selectedIssue.IssueNumber)
			return result, nil
		}

		fmt.Println("Selection out of range.")
	}
}
