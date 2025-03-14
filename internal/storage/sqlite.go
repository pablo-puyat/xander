package storage

import (
	"database/sql"
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
	// If dbPath is empty, use default location
	if dbPath == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine user home directory: %w", err)
		}
		configDir := filepath.Join(userHome, ".config", "xander")
		
		// Create directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}
		
		dbPath = filepath.Join(configDir, "xander.db")
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
	// Create comics table
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
	
	return nil
}

// StoreComic saves comic metadata to storage
func (s *SQLiteStorage) StoreComic(result *comicvine.Result) error {
	// Check if comic already exists
	var id int
	err := s.db.QueryRow("SELECT id FROM comics WHERE comicvine_id = ?", result.ComicVineID).Scan(&id)
	
	// If comic exists, update it
	if err == nil {
		_, err = s.db.Exec(`
		UPDATE comics SET 
			series = ?,
			issue = ?,
			year = ?,
			publisher = ?,
			title = ?,
			cover_url = ?,
			description = ?,
			filename = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE comicvine_id = ?`,
			result.Series,
			result.Issue,
			result.Year,
			result.Publisher,
			result.Title,
			result.CoverURL,
			result.Description,
			result.Filename,
			result.ComicVineID)
			
		if err != nil {
			return fmt.Errorf("failed to update comic: %w", err)
		}
		
		return nil
	}
	
	// If comic doesn't exist, insert it
	if err == sql.ErrNoRows {
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
			filename
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			result.ComicVineID,
			result.Series,
			result.Issue,
			result.Year,
			result.Publisher,
			result.Title,
			result.CoverURL,
			result.Description,
			result.Filename)
			
		if err != nil {
			return fmt.Errorf("failed to insert comic: %w", err)
		}
		
		return nil
	}
	
	// Some other error occurred
	return fmt.Errorf("failed to check if comic exists: %w", err)
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
