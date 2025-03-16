package comicvine

import (
	"fmt"
	"log"
	"strings"
	"xander/internal/comic"
)

type ComicClient interface {
	Get(series string) (*Series, error)
}

type ComicService struct {
	client  ComicClient
	verbose bool
}

func NewService(apiKey string, verbose bool) *ComicService {
	return &ComicService{
		client:  NewClient(apiKey, verbose),
		verbose: verbose,
	}
}

type Result struct {
	// File information
	Filename string

	// Basic metadata (from filename parsing)
	Series    string
	Issue     string
	Year      string
	Publisher string

	// Basic ComicVine data
	ComicVineID  int
	Title        string
	CoverURL     string
	Description  string
	ApiPublisher string // Publisher from API

	// Extended ComicVine data
	Volume          map[string]interface{}   // All volume information
	Characters      []map[string]interface{} // All character information
	Teams           []map[string]interface{} // All team information
	Locations       []map[string]interface{} // All location information
	Concepts        []map[string]interface{} // All concept information
	Objects         []map[string]interface{} // All object information
	People          []map[string]interface{} // All people credits information
	StoreDate       string
	CoverDate       string
	DateAdded       string
	DateLastUpdated string
	Image           map[string]interface{} // All image information

	// Full raw data
	RawData map[string]interface{} // Complete raw response
}

// GetMetadataWithInfo retrieves metadata for a comic with pre-parsed information
func (s *ComicService) GetMetadataWithInfo(series, issue, year string, filename string) (*Result, error) {
	if s.verbose {
		log.Printf("Processing comic with Series='%s', Issue='%s', Year='%s', Filename='%s'",
			series, issue, year, filename)
	}

	// Get issue from ComicVine API
	comicInfo, err := s.client.GetIssue(series, issue)
	if err != nil {
		if s.verbose {
			log.Printf("Failed to get issue from ComicVine for '%s' #%s: %v", series, issue, err)
		}
		return nil, fmt.Errorf("failed to get issue from ComicVine: %w", err)
	}

	if s.verbose {
		log.Printf("Successfully retrieved metadata for '%s' #%s", series, issue)
	}

	// Create the result
	result := &Result{
		// File info
		Filename:  filename,
		Series:    series,
		Issue:     issue,
		Year:      year,
		Publisher: comicInfo.Publisher,

		// Basic API data
		ComicVineID: comicInfo.ID,
		Title:       comicInfo.Name,
		Description: comicInfo.Description,

		// Extract data from maps
		RawData: comicInfo.RawData,
	}

	// Set image URL directly from the struct
	result.CoverURL = comicInfo.Image.OriginalURL

	// Extract additional data from raw data
	if comicInfo.RawData != nil {
		s.extractVolumeInfo(comicInfo.RawData, result, issue)
		s.extractCharacters(comicInfo.RawData, result)
		s.extractTeams(comicInfo.RawData, result)
		s.extractPeople(comicInfo.RawData, result)
		s.extractLocations(comicInfo.RawData, result)
		s.extractConcepts(comicInfo.RawData, result)
		s.extractObjects(comicInfo.RawData, result)
		s.extractDates(comicInfo.RawData, result)
		s.extractImage(comicInfo.RawData, result)
	}

	// Store other data directly from the Issue struct
	result.CoverDate = comicInfo.CoverDate
	result.StoreDate = comicInfo.StoreDate

	return result, nil
}

// Helper methods for extracting data from the raw response
func (s *ComicService) extractVolumeInfo(rawData map[string]interface{}, result *Result, issue string) {
	if volumeData, ok := rawData["volume"].(map[string]interface{}); ok {
		// Create a copy of the volume data to avoid sharing references
		volumeCopy := make(map[string]interface{})
		for k, v := range volumeData {
			volumeCopy[k] = v
		}

		// Add issue-specific metadata to avoid duplicate volume data
		volumeCopy["issue_id"] = result.ComicVineID
		volumeCopy["issue_number"] = issue

		// Verify volume data - if name doesn't match the series, this is likely incorrect
		if volumeName, ok := volumeCopy["name"].(string); ok {
			if !strings.Contains(strings.ToLower(volumeName), strings.ToLower(result.Series)) {
				// Log the mismatch but don't immediately abort - try to fix
				if s.verbose {
					log.Printf("Warning: Volume name '%s' doesn't match series '%s', fixing...",
						volumeName, result.Series)
				}
				// Override with series name to ensure correct data
				volumeCopy["name"] = result.Series

				// If we have ID data, check that too
				if volID, ok := volumeCopy["id"].(float64); ok {
					volumeCopy["original_id"] = volID // Keep original for reference
					// We don't have the correct ID, but ensure it's unique at least
					// Use a hash of the series name as a temporary ID
					h := 0
					for _, c := range result.Series {
						h = 31*h + int(c)
					}
					volumeCopy["id"] = h
				}
			}
		}

		result.Volume = volumeCopy

		// Extract publisher from volume (using original volume data)
		if pubData, ok := volumeData["publisher"].(map[string]interface{}); ok {
			if pubName, ok := pubData["name"].(string); ok {
				result.ApiPublisher = pubName
			}
		}
	}
}

func (s *ComicService) extractCharacters(rawData map[string]interface{}, result *Result) {
	if charData, ok := rawData["character_credits"].([]interface{}); ok {
		for _, char := range charData {
			if charMap, ok := char.(map[string]interface{}); ok {
				result.Characters = append(result.Characters, charMap)
			}
		}
	}
}

func (s *ComicService) extractTeams(rawData map[string]interface{}, result *Result) {
	if teamData, ok := rawData["team_credits"].([]interface{}); ok {
		for _, team := range teamData {
			if teamMap, ok := team.(map[string]interface{}); ok {
				result.Teams = append(result.Teams, teamMap)
			}
		}
	}
}

func (s *ComicService) extractPeople(rawData map[string]interface{}, result *Result) {
	if peopleData, ok := rawData["person_credits"].([]interface{}); ok {
		for _, person := range peopleData {
			if personMap, ok := person.(map[string]interface{}); ok {
				result.People = append(result.People, personMap)
			}
		}
	}
}

func (s *ComicService) extractLocations(rawData map[string]interface{}, result *Result) {
	if locData, ok := rawData["location_credits"].([]interface{}); ok {
		for _, loc := range locData {
			if locMap, ok := loc.(map[string]interface{}); ok {
				result.Locations = append(result.Locations, locMap)
			}
		}
	}
}

func (s *ComicService) extractConcepts(rawData map[string]interface{}, result *Result) {
	if conceptData, ok := rawData["concept_credits"].([]interface{}); ok {
		for _, concept := range conceptData {
			if conceptMap, ok := concept.(map[string]interface{}); ok {
				result.Concepts = append(result.Concepts, conceptMap)
			}
		}
	}
}

func (s *ComicService) extractObjects(rawData map[string]interface{}, result *Result) {
	if objData, ok := rawData["object_credits"].([]interface{}); ok {
		for _, obj := range objData {
			if objMap, ok := obj.(map[string]interface{}); ok {
				result.Objects = append(result.Objects, objMap)
			}
		}
	}
}

func (s *ComicService) extractDates(rawData map[string]interface{}, result *Result) {
	if dateAdded, ok := rawData["date_added"].(string); ok {
		result.DateAdded = dateAdded
	}

	if dateUpdated, ok := rawData["date_last_updated"].(string); ok {
		result.DateLastUpdated = dateUpdated
	}
}

func (s *ComicService) extractImage(rawData map[string]interface{}, result *Result) {
	if imgData, ok := rawData["image"].(map[string]interface{}); ok {
		result.Image = imgData
	}
}

// GetMetadata is a placeholder that should be implemented by the user
// to integrate with their parser to extract metadata from filenames
func (s *ComicService) GetMetadata(filename string) (*Result, error) {
	return nil, fmt.Errorf("method not implemented - needs external filename parser")
}

// GetMetadataForFiles processes multiple files and returns their metadata
func (s *ComicService) GetMetadataForFiles(filenames []string) ([]*Result, error) {
	var results []*Result
	var errors []string

	total := len(filenames)

	for i, filename := range filenames {
		// Show progress before processing each file
		fmt.Printf("Processing %d of %d: %s\n", i+1, total, filename)

		result, err := s.GetMetadata(filename)
		if err != nil {
			// Log error but continue with other files
			errorMsg := fmt.Sprintf("Error processing %s: %v", filename, err)
			fmt.Println(errorMsg)
			errors = append(errors, errorMsg)
			continue
		}

		// Show success message
		fmt.Printf("✓ Found metadata for %s (Series: %s, Issue: %s)\n",
			filename, result.Series, result.Issue)

		results = append(results, result)
	}

	fmt.Printf("\nSuccessfully processed %d of %d files\n\n", len(results), total)

	// Only return error if no results were found
	if len(results) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("failed to process any files: %s", strings.Join(errors, "; "))
	}

	return results, nil
}

// ToComic converts a ComicVine API Result to a domain Comic model
func (r *Result) ToComic() *comic.Comic {
	// Get publisher from API data if available, fallback to parsed publisher
	publisher := r.Publisher
	if r.ApiPublisher != "" {
		publisher = r.ApiPublisher
	}

	return &comic.Comic{
		// Basic file/parsed info
		Filename:  r.Filename,
		Series:    r.Series,
		Issue:     r.Issue,
		Year:      r.Year,
		Publisher: publisher,

		// Basic API data
		ComicVineID: r.ComicVineID,
		Title:       r.Title,
		CoverURL:    r.CoverURL,
		Description: r.Description,

		// Extended API data
		Volume:          r.Volume,
		Characters:      r.Characters,
		Teams:           r.Teams,
		Locations:       r.Locations,
		Concepts:        r.Concepts,
		Objects:         r.Objects,
		People:          r.People,
		StoreDate:       r.StoreDate,
		CoverDate:       r.CoverDate,
		DateAdded:       r.DateAdded,
		DateLastUpdated: r.DateLastUpdated,
		Image:           r.Image,

		// Complete raw data
		RawData: r.RawData,
	}
}

// FromComic creates a ComicVine Result from a domain Comic model
func FromComic(c *comic.Comic) *Result {
	return &Result{
		// Basic info
		Filename:  c.Filename,
		Series:    c.Series,
		Issue:     c.Issue,
		Year:      c.Year,
		Publisher: c.Publisher,

		// Basic API data
		ComicVineID: c.ComicVineID,
		Title:       c.Title,
		CoverURL:    c.CoverURL,
		Description: c.Description,

		// Extended data
		Volume:          c.Volume,
		Characters:      c.Characters,
		Teams:           c.Teams,
		Locations:       c.Locations,
		Concepts:        c.Concepts,
		Objects:         c.Objects,
		People:          c.People,
		StoreDate:       c.StoreDate,
		CoverDate:       c.CoverDate,
		DateAdded:       c.DateAdded,
		DateLastUpdated: c.DateLastUpdated,
		Image:           c.Image,

		// Raw data
		RawData: c.RawData,
	}
}
