package tui

import (
	"context"
	"fmt"

	"comic-parser/internal/comicvine"
	"comic-parser/internal/models"
	"comic-parser/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	store    *storage.Storage
	cvClient *comicvine.Client
	items    []*models.ParsedFilename
	index    int

	searchResults []models.ComicVineIssue
	searching     bool
	searchErr     error

	width  int
	height int
}

func NewModel(store *storage.Storage, cvClient *comicvine.Client) (Model, error) {
	// Load items initially
	items, err := store.ListParsedFilenames(context.Background())
	if err != nil {
		return Model{}, err
	}

	return Model{
		store:    store,
		cvClient: cvClient,
		items:    items,
		index:    0,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

type searchMsg struct {
	results []models.ComicVineIssue
	err     error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "n", "right", "l":
			if m.index < len(m.items)-1 {
				m.index++
				m.searchResults = nil // Clear search results on nav
				m.searchErr = nil
			}
		case "p", "left", "h":
			if m.index > 0 {
				m.index--
				m.searchResults = nil
				m.searchErr = nil
			}
		case "s", "enter": // Search
			if !m.searching && len(m.items) > 0 {
				m.searching = true
				m.searchResults = nil
				m.searchErr = nil
				item := m.items[m.index]
				return m, func() tea.Msg {
					// Use a new context?
					results, err := m.cvClient.SearchIssues(context.Background(), item.Title, item.IssueNumber)
					return searchMsg{results: results, err: err}
				}
			}
		}

	case searchMsg:
		m.searching = false
		if msg.err != nil {
			m.searchErr = msg.err
		} else {
			m.searchResults = msg.results
		}
	}

	return m, nil
}

func (m Model) View() string {
	if len(m.items) == 0 {
		return "No items found in database.\n\nPress 'q' to quit."
	}

	item := m.items[m.index]

	s := fmt.Sprintf("Item %d of %d\n\n", m.index+1, len(m.items))
	s += fmt.Sprintf("Filename: %s\n", item.OriginalFilename)
	s += fmt.Sprintf("Title:    %s\n", item.Title)
	s += fmt.Sprintf("Issue:    %s\n", item.IssueNumber)
	s += fmt.Sprintf("Year:     %s\n", item.Year)
	s += fmt.Sprintf("Conf:     %s\n", item.Confidence)
	if item.Notes != "" {
		s += fmt.Sprintf("Notes:    %s\n", item.Notes)
	}

	s += "\n---\n"

	if m.searching {
		s += "Searching ComicVine...\n"
	} else if m.searchErr != nil {
		s += fmt.Sprintf("Error: %v\n", m.searchErr)
	} else if m.searchResults != nil && len(m.searchResults) > 0 {
		s += fmt.Sprintf("Found %d matches:\n", len(m.searchResults))
		for i, res := range m.searchResults {
			if i >= 5 { // Limit to 5
				s += fmt.Sprintf("... and %d more\n", len(m.searchResults)-5)
				break
			}
			s += fmt.Sprintf("- %s #%s (%s) [%d]\n", res.Volume.Name, res.IssueNumber, res.CoverDate, res.ID)
		}
	} else if m.searchResults != nil { // empty but not nil means searched and found nothing
		s += "No matches found on ComicVine.\n"
	} else {
		s += "Press 's' or 'enter' to search ComicVine.\n"
	}

	s += "\n(n)ext, (p)rev, (s)earch, (q)uit\n"

	return s
}
