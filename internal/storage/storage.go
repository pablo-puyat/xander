package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"comic-parser/internal/db"
	"comic-parser/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS comic_vine_volumes (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    start_year TEXT,
    publisher_name TEXT,
    site_detail_url TEXT
);

CREATE TABLE IF NOT EXISTS comic_vine_issues (
    id INTEGER PRIMARY KEY,
    volume_id INTEGER NOT NULL,
    name TEXT,
    issue_number TEXT,
    cover_date TEXT,
    store_date TEXT,
    description TEXT,
    site_detail_url TEXT,
    image_small_url TEXT,
    image_medium_url TEXT,
    image_large_url TEXT,
    FOREIGN KEY (volume_id) REFERENCES comic_vine_volumes(id)
);

CREATE TABLE IF NOT EXISTS processing_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL UNIQUE,
    success BOOLEAN NOT NULL,
    error TEXT,
    processed_at DATETIME NOT NULL,
    processing_time_ms INTEGER NOT NULL,
    match_confidence TEXT,
    reasoning TEXT,
    comicvine_id INTEGER,
    comicvine_url TEXT,
    FOREIGN KEY (comicvine_id) REFERENCES comic_vine_issues(id)
);

CREATE TABLE IF NOT EXISTS parsed_filenames (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    processing_result_id INTEGER,
    parser_name TEXT NOT NULL DEFAULT 'unknown',
    original_filename TEXT NOT NULL,
    title TEXT NOT NULL,
    issue_number TEXT NOT NULL,
    year TEXT,
    publisher TEXT,
    volume_number TEXT,
    confidence TEXT NOT NULL,
    notes TEXT,
    FOREIGN KEY (processing_result_id) REFERENCES processing_results(id) ON DELETE CASCADE,
    UNIQUE(original_filename, parser_name)
);
`

type Storage struct {
	db *sql.DB
	q  *db.Queries
}

func NewStorage(dbPath string) (*Storage, error) {
	dbConn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := dbConn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err := dbConn.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create tables
	if _, err := dbConn.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &Storage{
		db: dbConn,
		q:  db.New(dbConn),
	}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveResult(ctx context.Context, result *models.ProcessingResult) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	// Save ComicVine data if match exists
	var cvID sql.NullInt64
	var cvURL sql.NullString

	if result.Match != nil && result.Match.SelectedIssue != nil {
		issue := result.Match.SelectedIssue
		vol := issue.Volume

		// Save Volume
		err = qtx.UpsertVolume(ctx, db.UpsertVolumeParams{
			ID:            int64(vol.ID),
			Name:          vol.Name,
			StartYear:     sql.NullString{}, // Not in VolumeRef
			PublisherName: sql.NullString{String: vol.Publisher, Valid: vol.Publisher != ""},
			SiteDetailUrl: sql.NullString{String: vol.SiteURL, Valid: vol.SiteURL != ""},
		})
		if err != nil {
			return fmt.Errorf("failed to upsert volume: %w", err)
		}

		// Save Issue
		err = qtx.UpsertIssue(ctx, db.UpsertIssueParams{
			ID:             int64(issue.ID),
			VolumeID:       int64(vol.ID),
			Name:           sql.NullString{String: issue.Name, Valid: issue.Name != ""},
			IssueNumber:    sql.NullString{String: issue.IssueNumber, Valid: issue.IssueNumber != ""},
			CoverDate:      sql.NullString{String: issue.CoverDate, Valid: issue.CoverDate != ""},
			StoreDate:      sql.NullString{String: issue.StoreDate, Valid: issue.StoreDate != ""},
			Description:    sql.NullString{String: issue.Description, Valid: issue.Description != ""},
			SiteDetailUrl:  sql.NullString{String: issue.SiteDetailURL, Valid: issue.SiteDetailURL != ""},
			ImageSmallUrl:  sql.NullString{String: issue.Image.SmallURL, Valid: issue.Image.SmallURL != ""},
			ImageMediumUrl: sql.NullString{String: issue.Image.MediumURL, Valid: issue.Image.MediumURL != ""},
			ImageLargeUrl:  sql.NullString{String: issue.Image.LargeURL, Valid: issue.Image.LargeURL != ""},
		})
		if err != nil {
			return fmt.Errorf("failed to upsert issue: %w", err)
		}

		cvID = sql.NullInt64{Int64: int64(issue.ID), Valid: true}
		cvURL = sql.NullString{String: issue.SiteDetailURL, Valid: true}
	}

	// Save Processing Result
	matchConf := sql.NullString{}
	reasoning := sql.NullString{}

	if result.Match != nil {
		matchConf = sql.NullString{String: result.Match.MatchConfidence, Valid: true}
		reasoning = sql.NullString{String: result.Match.Reasoning, Valid: true}
	}

	// ProcessedAt is required, but if it's zero, we should probably set it to now
	processedAt := result.ProcessedAt
	if processedAt.IsZero() {
		processedAt = time.Now()
	}

	resID, err := qtx.UpsertProcessingResult(ctx, db.UpsertProcessingResultParams{
		Filename:         result.Filename,
		Success:          result.Success,
		Error:            sql.NullString{String: result.Error, Valid: result.Error != ""},
		ProcessedAt:      processedAt,
		ProcessingTimeMs: result.ProcessingTimeMS,
		MatchConfidence:  matchConf,
		Reasoning:        reasoning,
		ComicvineID:      cvID,
		ComicvineUrl:     cvURL,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert processing result: %w", err)
	}

	// Delete old parsed filenames
	if err := qtx.DeleteParsedFilenamesByResultID(ctx, resID); err != nil {
		return fmt.Errorf("failed to delete old parsed filenames: %w", err)
	}

	// Insert new parsed filename
	if result.Match != nil {
		info := result.Match.ParsedInfo
		err = qtx.CreateParsedFilename(ctx, db.CreateParsedFilenameParams{
			ProcessingResultID: sql.NullInt64{Int64: resID, Valid: true},
			ParserName:         "pipeline",
			OriginalFilename:   info.OriginalFilename,
			Title:              info.Title,
			IssueNumber:        info.IssueNumber,
			Year:               sql.NullString{String: info.Year, Valid: info.Year != ""},
			Publisher:          sql.NullString{String: info.Publisher, Valid: info.Publisher != ""},
			VolumeNumber:       sql.NullString{String: info.VolumeNumber, Valid: info.VolumeNumber != ""},
			Confidence:         info.Confidence,
			Notes:              sql.NullString{String: info.Notes, Valid: info.Notes != ""},
		})
		if err != nil {
			return fmt.Errorf("failed to create parsed filename: %w", err)
		}
	}

	return tx.Commit()
}

func (s *Storage) SaveParsedFilename(ctx context.Context, info *models.ParsedFilename, parserName string) error {
	return s.q.CreateParsedFilename(ctx, db.CreateParsedFilenameParams{
		ProcessingResultID: sql.NullInt64{Valid: false},
		ParserName:         parserName,
		OriginalFilename:   info.OriginalFilename,
		Title:              info.Title,
		IssueNumber:        info.IssueNumber,
		Year:               sql.NullString{String: info.Year, Valid: info.Year != ""},
		Publisher:          sql.NullString{String: info.Publisher, Valid: info.Publisher != ""},
		VolumeNumber:       sql.NullString{String: info.VolumeNumber, Valid: info.VolumeNumber != ""},
		Confidence:         info.Confidence,
		Notes:              sql.NullString{String: info.Notes, Valid: info.Notes != ""},
	})
}

func (s *Storage) ListParsedFilenames(ctx context.Context) ([]*models.ParsedFilename, error) {
	dbItems, err := s.q.ListParsedFilenames(ctx)
	if err != nil {
		return nil, err
	}

	var items []*models.ParsedFilename
	for _, dbItem := range dbItems {
		item := &models.ParsedFilename{
			OriginalFilename: dbItem.OriginalFilename,
			Title:            dbItem.Title,
			IssueNumber:      dbItem.IssueNumber,
			Year:             dbItem.Year.String,
			Publisher:        dbItem.Publisher.String,
			VolumeNumber:     dbItem.VolumeNumber.String,
			Confidence:       dbItem.Confidence,
			Notes:            dbItem.Notes.String,
		}
		items = append(items, item)
	}
	return items, nil
}
