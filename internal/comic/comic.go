package comic

// Comic represents a comic book in the application domain
type Comic struct {
	Filename    string
	Series      string
	Issue       string
	Year        string
	Publisher   string
	ComicVineID int
	Title       string
	CoverURL    string
	Description string
}

// Filter defines criteria for filtering comics
type Filter struct {
	Series    string
	Issue     string
	Year      string
	Publisher string
	Filename  string
}

// Storage defines the interface for comic persistence
type Storage interface {
	StoreComic(comic *Comic) error
	Close() error
}
