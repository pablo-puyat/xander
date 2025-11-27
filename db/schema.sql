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
    processing_result_id INTEGER NOT NULL,
    original_filename TEXT NOT NULL,
    title TEXT NOT NULL,
    issue_number TEXT NOT NULL,
    year TEXT,
    publisher TEXT,
    volume_number TEXT,
    confidence TEXT NOT NULL,
    notes TEXT,
    FOREIGN KEY (processing_result_id) REFERENCES processing_results(id) ON DELETE CASCADE
);
