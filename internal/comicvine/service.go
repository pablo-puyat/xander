package comicvine

import (
	"fmt"
	"log"
	"strings"
	"xander/internal/comic"
	"xander/internal/parse"
)

// ComicService represents a service for comic metadata operations
type ComicService struct {
	client *Client
	verbose bool
}

// NewComicService creates a new comic service
func NewComicService(apiKey string, verbose bool) *ComicService {
	return &ComicService{
		client: NewClient(apiKey, verbose),
		verbose: verbose,
	}
}

// Result represents the metadata result for a comic file
type Result struct {
	// File information
	Filename    string
	
	// Basic metadata (from filename parsing)
	Series      string
	Issue       string
	Year        string
	Publisher   string
	
	// Basic ComicVine data
	ComicVineID int
	Title       string
	CoverURL    string
	Description string
	ApiPublisher string // Publisher from API
	
	// Extended ComicVine data
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
	
	// Full raw data
	RawData            map[string]interface{} // Complete raw response
}

// GetMetadata retrieves metadata for a comic file
func (s *ComicService) GetMetadata(filename string) (*Result, error) {
	// Extract info from the filename
	series, issue, year, publisher, err := parse.ParseComicFilename(filename)
	if err != nil {
		if s.verbose {
			log.Printf("Failed to parse filename '%s': %v", filename, err)
		}
		return nil, fmt.Errorf("failed to parse filename: %w", err)
	}

	if s.verbose {
		log.Printf("Parsed '%s' as Series='%s', Issue='%s', Year='%s', Publisher='%s'", 
			filename, series, issue, year, publisher)
	}

	// Get issue from ComicVine
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
		Filename:    filename,
		Series:      series,
		Issue:       issue,
		Year:        year,
		Publisher:   publisher,
		
		// Basic API data
		ComicVineID: comicInfo.ID,
		Title:       comicInfo.Name,
		Description: comicInfo.Description,
		
		// Extract data from maps
		RawData:     comicInfo.RawData,
	}
	
	// Extract all the data from the API response
	
	// Set image URL directly from the struct
	result.CoverURL = comicInfo.Image.OriginalURL
	
	// Extract additional data from raw data
	if comicInfo.RawData != nil {
		// Extract volume info
		if volumeData, ok := comicInfo.RawData["volume"].(map[string]interface{}); ok {
			result.Volume = volumeData
			
			// Extract publisher from volume
			if pubData, ok := volumeData["publisher"].(map[string]interface{}); ok {
				if pubName, ok := pubData["name"].(string); ok {
					result.ApiPublisher = pubName
				}
			}
		}
		
		// Extract character data
		if charData, ok := comicInfo.RawData["character_credits"].([]interface{}); ok {
			for _, char := range charData {
				if charMap, ok := char.(map[string]interface{}); ok {
					result.Characters = append(result.Characters, charMap)
				}
			}
		}
		
		// Extract team data
		if teamData, ok := comicInfo.RawData["team_credits"].([]interface{}); ok {
			for _, team := range teamData {
				if teamMap, ok := team.(map[string]interface{}); ok {
					result.Teams = append(result.Teams, teamMap)
				}
			}
		}
		
		// Extract people data
		if peopleData, ok := comicInfo.RawData["person_credits"].([]interface{}); ok {
			for _, person := range peopleData {
				if personMap, ok := person.(map[string]interface{}); ok {
					result.People = append(result.People, personMap)
				}
			}
		}
		
		// Extract location data
		if locData, ok := comicInfo.RawData["location_credits"].([]interface{}); ok {
			for _, loc := range locData {
				if locMap, ok := loc.(map[string]interface{}); ok {
					result.Locations = append(result.Locations, locMap)
				}
			}
		}
		
		// Extract concept data
		if conceptData, ok := comicInfo.RawData["concept_credits"].([]interface{}); ok {
			for _, concept := range conceptData {
				if conceptMap, ok := concept.(map[string]interface{}); ok {
					result.Concepts = append(result.Concepts, conceptMap)
				}
			}
		}
		
		// Extract object data
		if objData, ok := comicInfo.RawData["object_credits"].([]interface{}); ok {
			for _, obj := range objData {
				if objMap, ok := obj.(map[string]interface{}); ok {
					result.Objects = append(result.Objects, objMap)
				}
			}
		}
		
		// Extract dates
		if dateAdded, ok := comicInfo.RawData["date_added"].(string); ok {
			result.DateAdded = dateAdded
		}
		
		if dateUpdated, ok := comicInfo.RawData["date_last_updated"].(string); ok {
			result.DateLastUpdated = dateUpdated
		}
		
		// Store complete image data
		if imgData, ok := comicInfo.RawData["image"].(map[string]interface{}); ok {
			result.Image = imgData
		}
	}
	
	// Store other data directly from the Issue struct
	result.CoverDate = comicInfo.CoverDate
	result.StoreDate = comicInfo.StoreDate
	
	// The remaining data should already be set above through the RawData extraction
	// This avoids double-assignment

	return result, nil
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
		Filename:    r.Filename,
		Series:      r.Series,
		Issue:       r.Issue,
		Year:        r.Year,
		Publisher:   publisher,
		
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
		RawData:         r.RawData,
	}
}

// FromComic creates a ComicVine Result from a domain Comic model
func FromComic(c *comic.Comic) *Result {
	return &Result{
		// Basic info
		Filename:    c.Filename,
		Series:      c.Series,
		Issue:       c.Issue,
		Year:        c.Year,
		Publisher:   c.Publisher,
		
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
		RawData:         c.RawData,
	}
}