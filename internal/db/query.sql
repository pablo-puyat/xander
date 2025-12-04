-- name: UpsertVolume :exec
INSERT INTO comic_vine_volumes (
    id, name, start_year, publisher_name, site_detail_url
) VALUES (
    ?, ?, ?, ?, ?
) ON CONFLICT(id) DO UPDATE SET
    name = excluded.name,
    start_year = excluded.start_year,
    publisher_name = excluded.publisher_name,
    site_detail_url = excluded.site_detail_url;

-- name: UpsertIssue :exec
INSERT INTO comic_vine_issues (
    id, volume_id, name, issue_number, cover_date, store_date, description,
    site_detail_url, image_small_url, image_medium_url, image_large_url
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) ON CONFLICT(id) DO UPDATE SET
    volume_id = excluded.volume_id,
    name = excluded.name,
    issue_number = excluded.issue_number,
    cover_date = excluded.cover_date,
    store_date = excluded.store_date,
    description = excluded.description,
    site_detail_url = excluded.site_detail_url,
    image_small_url = excluded.image_small_url,
    image_medium_url = excluded.image_medium_url,
    image_large_url = excluded.image_large_url;

-- name: UpsertProcessingResult :one
INSERT INTO processing_results (
    filename, success, error, processed_at, processing_time_ms,
    match_confidence, reasoning, comicvine_id, comicvine_url
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
) ON CONFLICT(filename) DO UPDATE SET
    success = excluded.success,
    error = excluded.error,
    processed_at = excluded.processed_at,
    processing_time_ms = excluded.processing_time_ms,
    match_confidence = excluded.match_confidence,
    reasoning = excluded.reasoning,
    comicvine_id = excluded.comicvine_id,
    comicvine_url = excluded.comicvine_url
RETURNING id;

-- name: DeleteParsedFilenamesByResultID :exec
DELETE FROM parsed_filenames WHERE processing_result_id = ?;

-- name: CreateParsedFilename :exec
INSERT INTO parsed_filenames (
    processing_result_id, parser_name, original_filename, title, issue_number, year,
    publisher, volume_number, confidence, notes
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) ON CONFLICT(original_filename, parser_name) DO UPDATE SET
    processing_result_id = excluded.processing_result_id,
    title = excluded.title,
    issue_number = excluded.issue_number,
    year = excluded.year,
    publisher = excluded.publisher,
    volume_number = excluded.volume_number,
    confidence = excluded.confidence,
    notes = excluded.notes;

-- name: GetProcessingResult :one
SELECT * FROM processing_results WHERE filename = ?;

-- name: ListParsedFilenames :many
SELECT * FROM parsed_filenames ORDER BY id DESC;
