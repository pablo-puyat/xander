package tui

import (
	"context"
	"fmt"
	"strings"

	"comic-parser/internal/comicvine"
	"comic-parser/internal/models"
	"comic-parser/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

const maxSearchResults = 5

type Model struct {
	ctx      context.Context
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

func NewModel(ctx context.Context, store *storage.Storage, cvClient *comicvine.Client) (Model, error) {
	// Load items initially
	items, err := store.ListParsedFilenames(context.Background())
	if err != nil {
		return Model{}, err
	}

	return Model{
		ctx:      ctx,
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
	id      string
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
			m.navigate(1)
		case "p", "left", "h":
			m.navigate(-1)
		case "s", "enter": // Search
			if !m.searching && len(m.items) > 0 {
				m.searching = true
				m.searchResults = nil
				m.searchErr = nil
				item := m.items[m.index]
				return m, func() tea.Msg {
					results, err := m.cvClient.SearchIssues(m.ctx, item.Title, item.IssueNumber)
					return searchMsg{id: item.OriginalFilename, results: results, err: err}
				}
			}
		}

	case searchMsg:
		if m.index < len(m.items) && m.items[m.index].OriginalFilename == msg.id {
			m.searching = false
			if msg.err != nil {
				m.searchErr = msg.err
			} else {
				m.searchResults = msg.results
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if len(m.items) == 0 {
		return "No items found in database.\n\nPress 'q' to quit."
	}

	var b strings.Builder

	item := m.items[m.index]

	// 3. Write directly to the builder using Fprintf
	fmt.Fprintf(&b, "Item %d of %d\n\n", m.index+1, len(m.items))
	fmt.Fprintf(&b, "Filename: %s\n", item.OriginalFilename)
	fmt.Fprintf(&b, "Title:    %s\n", item.Title)
	fmt.Fprintf(&b, "Issue:    %s\n", item.IssueNumber)
	fmt.Fprintf(&b, "Year:     %s\n", item.Year)
	fmt.Fprintf(&b, "Conf:     %s\n", item.Confidence)

	if item.Notes != "" {
		fmt.Fprintf(&b, "Notes:    %s\n", item.Notes)
	}

	b.WriteString("\n---\n")

	if m.searching {
		b.WriteString("Searching ComicVine...\n")
	} else if m.searchErr != nil {
		fmt.Fprintf(&b, "Error: %v\n", m.searchErr)
	} else if len(m.searchResults) > 0 {
		fmt.Fprintf(&b, "Found %d matches:\n", len(m.searchResults))
		for i, res := range m.searchResults {
			if i >= 5 {
				fmt.Fprintf(&b, "... and %d more\n", len(m.searchResults)-maxSearchResults)
				break
			}
			fmt.Fprintf(&b, "- %s #%s (%s) [%d]\n", res.Volume.Name, res.IssueNumber, res.CoverDate, res.ID)
		}
	} else if m.searchResults != nil {
		b.WriteString("No matches found on ComicVine.\n")
	} else {
		b.WriteString("Press 's' or 'enter' to search ComicVine.\n")
	}

	b.WriteString("\n(n)ext, (p)rev, (s)earch, (q)uit\n")

	return b.String()
}

func (m *Model) navigate(offset int) {
	newIndex := m.index + offset
	if newIndex >= 0 && newIndex < len(m.items) {
		m.index = newIndex
		m.searchResults = nil
		m.searchErr = nil
	}
}
