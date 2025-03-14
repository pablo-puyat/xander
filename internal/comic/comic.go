package comic

// Comic represents a comic book in the application domain
type Comic struct {
	// File information
	Filename    string
	
	// Basic metadata
	Series      string
	Issue       string
	Year        string
	Publisher   string
	
	// ComicVine data
	ComicVineID int
	Title       string
	CoverURL    string
	Description string
	
	// Extended metadata
	Volume             map[string]interface{} // All volume information
	Characters         []map[string]interface{} // All character information
	Teams              []map[string]interface{} // All team information
	Locations          []map[string]interface{} // All location information
	Concepts           []map[string]interface{} // All concept information
	Objects            []map[string]interface{} // All object information
	People             []map[string]interface{} // All people credits information
	StoreDate          string
	CoverDate          string
	DateAdded          string
	DateLastUpdated    string
	Image              map[string]interface{} // All image information
	
	// Full raw response data
	RawData            map[string]interface{} // Store complete raw response
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
