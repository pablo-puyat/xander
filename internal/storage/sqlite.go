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

// GetComics retrieves all stored comics
func (s *SQLiteStorage) GetComics() ([]*comicvine.Result, error) {
	filter := NewFilter()
	return s.GetComicsByFilter(filter)
}

// GetComicByID retrieves a specific comic by its ComicVine ID
func (s *SQLiteStorage) GetComicByID(id int) (*comicvine.Result, error) {
	var result comicvine.Result
	
	err := s.db.QueryRow(`
	SELECT 
		comicvine_id,
		series,
		issue,
		year,
		publisher,
		title,
		cover_url,
		description,
		filename
	FROM comics
	WHERE comicvine_id = ?`,
		id).Scan(
		&result.ComicVineID,
		&result.Series,
		&result.Issue,
		&result.Year,
		&result.Publisher,
		&result.Title,
		&result.CoverURL,
		&result.Description,
		&result.Filename)
		
	if err == sql.ErrNoRows {
		return nil, nil // Not found, but not an error
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get comic: %w", err)
	}
	
	return &result, nil
}

// GetComicsByFilter retrieves comics matching the provided filter criteria
func (s *SQLiteStorage) GetComicsByFilter(filter ComicFilter) ([]*comicvine.Result, error) {
	// Build query with filters
	query := `
	SELECT 
		comicvine_id,
		series,
		issue,
		year,
		publisher,
		title,
		cover_url,
		description,
		filename
	FROM comics
	WHERE 1=1`
	
	var args []interface{}
	
	// Add filter conditions
	if filter.Series != "" {
		query += " AND series LIKE ?"
		args = append(args, "%"+filter.Series+"%")
	}
	
	if filter.Issue != "" {
		query += " AND issue = ?"
		args = append(args, filter.Issue)
	}
	
	if filter.Year != "" {
		query += " AND year = ?"
		args = append(args, filter.Year)
	}
	
	if filter.Publisher != "" {
		query += " AND publisher LIKE ?"
		args = append(args, "%"+filter.Publisher+"%")
	}
	
	if filter.Filename != "" {
		query += " AND filename LIKE ?"
		args = append(args, "%"+filter.Filename+"%")
	}
	
	if !filter.StartDate.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filter.StartDate)
	}
	
	if !filter.EndDate.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filter.EndDate)
	}
	
	// Add limit and offset
	query += " ORDER BY series, issue LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)
	
	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query comics: %w", err)
	}
	defer rows.Close()
	
	// Process results
	var results []*comicvine.Result
	for rows.Next() {
		var result comicvine.Result
		err := rows.Scan(
			&result.ComicVineID,
			&result.Series,
			&result.Issue,
			&result.Year,
			&result.Publisher,
			&result.Title,
			&result.CoverURL,
			&result.Description,
			&result.Filename)
			
		if err != nil {
			return nil, fmt.Errorf("failed to scan comic: %w", err)
		}
		
		results = append(results, &result)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comic rows: %w", err)
	}
	
	return results, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}