package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"xander/internal/comicvine"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// SQLiteStorage implements the Storage interface using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// If dbPath is empty, use default location following XDG spec
	if dbPath == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine user home directory: %w", err)
		}
		
		// Use ~/.local/share/xander/ for data files like databases
		dataDir := filepath.Join(userHome, ".local", "share", "xander")
		
		// Create directory if it doesn't exist
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
		
		dbPath = filepath.Join(dataDir, "xander.db")
	}
	
	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}
	
	return &SQLiteStorage{db: db}, nil
}

// createTables creates the necessary database tables if they don't exist
func createTables(db *sql.DB) error {
	// Create comics table with extended fields
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS comics (
		id INTEGER PRIMARY KEY,
		comicvine_id INTEGER NOT NULL,
		series TEXT NOT NULL,
		issue TEXT NOT NULL,
		year TEXT,
		publisher TEXT,
		title TEXT,
		cover_url TEXT,
		description TEXT,
		filename TEXT,
		
		-- Extended metadata fields
		store_date TEXT,
		cover_date TEXT,
		date_added TEXT,
		date_last_updated TEXT,
		
		-- JSON fields to store complex data
		volume_json TEXT,          -- JSON string of volume data
		characters_json TEXT,      -- JSON string of character data
		teams_json TEXT,           -- JSON string of team data
		people_json TEXT,          -- JSON string of people credits
		locations_json TEXT,       -- JSON string of locations
		concepts_json TEXT,        -- JSON string of concepts
		objects_json TEXT,         -- JSON string of objects
		image_json TEXT,           -- JSON string of all image data
		raw_data_json TEXT,        -- Complete raw response JSON
		
		-- Timestamps
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	
	if err != nil {
		return fmt.Errorf("failed to create comics table: %w", err)
	}
	
	// Create index on comicvine_id
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_comics_comicvine_id ON comics(comicvine_id)`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	
	// Create index on filename for faster lookups
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_comics_filename ON comics(filename COLLATE NOCASE)`)
	if err != nil {
		return fmt.Errorf("failed to create filename index: %w", err)
	}
	
	return nil
}

// StoreComic saves comic metadata to storage
func (s *SQLiteStorage) StoreComic(result *comicvine.Result) error {
	// First check if the comic already exists by filename (only use filename for identification)
	var id int
	err := s.db.QueryRow("SELECT id FROM comics WHERE LOWER(filename) = LOWER(?)", 
		result.Filename).Scan(&id)
	
	// Convert complex data to JSON strings
	volumeJSON, _ := json.Marshal(result.Volume)
	charactersJSON, _ := json.Marshal(result.Characters)
	teamsJSON, _ := json.Marshal(result.Teams)
	peopleJSON, _ := json.Marshal(result.People)
	locationsJSON, _ := json.Marshal(result.Locations)
	conceptsJSON, _ := json.Marshal(result.Concepts)
	objectsJSON, _ := json.Marshal(result.Objects)
	imageJSON, _ := json.Marshal(result.Image)
	rawDataJSON, _ := json.Marshal(result.RawData)
	
	// If comic exists, update it
	if err == nil {
		fmt.Printf("Updating existing record for %s (ID: %d)\n", result.Filename, id)
		
		_, err = s.db.Exec(`
		UPDATE comics SET 
			comicvine_id = ?,
			series = ?,
			issue = ?,
			year = ?,
			publisher = ?,
			title = ?,
			cover_url = ?,
			description = ?,
			
			-- Extended metadata fields
			store_date = ?,
			cover_date = ?,
			date_added = ?,
			date_last_updated = ?,
			
			-- JSON data fields
			volume_json = ?,
			characters_json = ?,
			teams_json = ?,
			people_json = ?,
			locations_json = ?,
			concepts_json = ?,
			objects_json = ?,
			image_json = ?,
			raw_data_json = ?,
			
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
			result.ComicVineID,
			result.Series,
			result.Issue,
			result.Year,
			result.Publisher,
			result.Title,
			result.CoverURL,
			result.Description,
			
			// Extended fields
			result.StoreDate,
			result.CoverDate,
			result.DateAdded,
			result.DateLastUpdated,
			
			// JSON data
			string(volumeJSON),
			string(charactersJSON),
			string(teamsJSON),
			string(peopleJSON),
			string(locationsJSON),
			string(conceptsJSON),
			string(objectsJSON),
			string(imageJSON),
			string(rawDataJSON),
			
			id)
			
		if err != nil {
			return fmt.Errorf("failed to update comic: %w", err)
		}
		
		return nil
	}
	
	// If comic doesn't exist, insert it
	if err == sql.ErrNoRows {
		fmt.Printf("Inserting new record for %s\n", result.Filename)
		
		_, err = s.db.Exec(`
		INSERT INTO comics (
			comicvine_id,
			series,
			issue,
			year,
			publisher,
			title,
			cover_url,
			description,
			filename,
			
			-- Extended metadata fields
			store_date,
			cover_date,
			date_added,
			date_last_updated,
			
			-- JSON data fields
			volume_json,
			characters_json,
			teams_json,
			people_json,
			locations_json,
			concepts_json,
			objects_json,
			image_json,
			raw_data_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			result.ComicVineID,
			result.Series,
			result.Issue,
			result.Year,
			result.Publisher,
			result.Title,
			result.CoverURL,
			result.Description,
			result.Filename,
			
			// Extended fields
			result.StoreDate,
			result.CoverDate,
			result.DateAdded,
			result.DateLastUpdated,
			
			// JSON data
			string(volumeJSON),
			string(charactersJSON),
			string(teamsJSON),
			string(peopleJSON),
			string(locationsJSON),
			string(conceptsJSON),
			string(objectsJSON),
			string(imageJSON),
			string(rawDataJSON))
			
		if err != nil {
			return fmt.Errorf("failed to insert comic: %w", err)
		}
		
		return nil
	}
	
	// Some other error occurred
	return fmt.Errorf("failed to check if comic exists: %w", err)
}

// FilenameExistsInDb checks if a filename exists in the database
func (s *SQLiteStorage) FilenameExistsInDb(filename string) (bool, error) {
	// Just check if the exact filename exists in the database
	// This is used BEFORE parsing, so we only have the filename to work with
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM comics WHERE LOWER(filename) = LOWER(?)", filename).Scan(&count)
	
	if err != nil {
		return false, fmt.Errorf("failed to check for existing comic: %w", err)
	}
	
	if count > 0 {
		// Debug print to verify the file was found
		var storedFilename string
		err := s.db.QueryRow("SELECT filename FROM comics WHERE LOWER(filename) = LOWER(?)", filename).Scan(&storedFilename)
		if err == nil {
			fmt.Printf("Found in database with filename: %s\n", storedFilename)
		}
	}
	
	return count > 0, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}