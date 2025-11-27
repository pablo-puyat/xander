# Comic File Parser

A Go application that uses Claude (Anthropic's LLM) to parse comic book archive filenames and match them against ComicVine's database.

## Features

- **LLM-powered filename parsing**: Handles diverse filename formats like:
  - `Amazing Spider-Man 001 (2018).cbz`
  - `Batman - The Long Halloween 01.cbr`
  - `X-Men v2 #45 (1995).cbz`
  - `The Walking Dead #100 (2012) (Digital).cbz`
  
- **ComicVine API integration**: Searches for matching issues and volumes
- **LLM-powered result matching**: Intelligently selects the best match from multiple results
- **Batch processing**: Process thousands of files with configurable concurrency
- **Multiple output formats**: JSON and CSV export

## Prerequisites

1. **Anthropic API Key**: Get one from https://console.anthropic.com/
2. **ComicVine API Key**: Get one from https://comicvine.gamespot.com/api/

## Installation

```bash
# Clone or copy the project
cd comic-parser

# Build
go build -o comic-parser .
```

## Configuration

### Option 1: Environment Variables (Recommended)

```bash
export ANTHROPIC_API_KEY="your-anthropic-key"
export COMICVINE_API_KEY="your-comicvine-key"
```

### Option 2: Configuration File

Generate a sample config:
```bash
./comic-parser -generate-config
```

Then edit `config.json`:
```json
{
  "anthropic_api_key": "your-anthropic-key",
  "comicvine_api_key": "your-comicvine-key",
  "anthropic_model": "claude-sonnet-4-20250514",
  "worker_count": 3,
  "rate_limit_per_min": 30,
  "output_format": "json",
  "verbose": false
}
```

## Usage

### Process a Single File (Testing)

```bash
./comic-parser -file "Amazing Spider-Man 001 (2018).cbz" -verbose
```

Output:
```
Processing: Amazing Spider-Man 001 (2018).cbz

=== Parsed Information ===
Title:        Amazing Spider-Man
Issue:        1
Year:         2018
Publisher:    
Volume:       
Confidence:   high
Notes:        

=== ComicVine Match ===
Series:       The Amazing Spider-Man
Issue:        #1
Cover Date:   2018-07-01
Publisher:    Marvel
ComicVine ID: 123456
URL:          https://comicvine.gamespot.com/the-amazing-spider-man-1-back-to-basics-part-1/...
Confidence:   high
Reasoning:    Title matches, issue number matches, year aligns with cover date
```

### Batch Processing

Create a text file with filenames (one per line):

```bash
# filenames.txt
Amazing Spider-Man 001 (2018).cbz
Batman #50 (2018).cbr
X-Men v2 #45 (1995).cbz
```

Process the batch:
```bash
./comic-parser -input filenames.txt -output results.json -workers 3
```

### Command Line Options

```
Usage of comic-parser:
  -config string
        Path to configuration file (default "config.json")
  -file string
        Process a single filename (for testing)
  -format string
        Output format: json or csv (default "json")
  -generate-config
        Generate a sample config file
  -input string
        Input file containing filenames (one per line)
  -output string
        Output file for results (default "results.json")
  -verbose
        Enable verbose logging
  -workers int
        Number of concurrent workers (default 3)
```

## Output Format

### JSON Output

```json
[
  {
    "filename": "Amazing Spider-Man 001 (2018).cbz",
    "success": true,
    "match": {
      "original_filename": "Amazing Spider-Man 001 (2018).cbz",
      "parsed_info": {
        "title": "Amazing Spider-Man",
        "issue_number": "1",
        "year": "2018",
        "confidence": "high"
      },
      "selected_issue": {
        "id": 123456,
        "issue_number": "1",
        "cover_date": "2018-07-01",
        "volume": {
          "id": 789,
          "name": "The Amazing Spider-Man",
          "publisher": "Marvel"
        }
      },
      "match_confidence": "high",
      "reasoning": "Title and issue match exactly, year aligns",
      "comicvine_id": 123456,
      "comicvine_url": "https://comicvine.gamespot.com/..."
    },
    "processed_at": "2024-01-15T10:30:00Z",
    "processing_time_ms": 2500
  }
]
```

### CSV Output

Use `-format csv` for spreadsheet-compatible output.

## Rate Limiting

The application respects rate limits for both APIs:
- **Anthropic**: Configurable via `rate_limit_per_min` (default: 30/min)
- **ComicVine**: Built-in ~1 request/second limit

Adjust `worker_count` to balance speed vs. rate limits.

## Generating Input File

To generate a list of comic files from a directory:

```bash
# Linux/Mac
find /path/to/comics -name "*.cb?" -type f -exec basename {} \; > filenames.txt

# Windows PowerShell
Get-ChildItem -Path "C:\Comics" -Filter "*.cb?" -Recurse | ForEach-Object { $_.Name } > filenames.txt
```

## Handling Large Batches (11,000+ files)

For very large batches:

1. **Split into chunks**: Process 500-1000 files at a time
2. **Use conservative workers**: Start with 2-3 workers
3. **Monitor costs**: Each file requires 2 LLM calls
4. **Save progress**: Results are saved incrementally

```bash
# Split files
split -l 500 filenames.txt batch_

# Process each batch
for batch in batch_*; do
  ./comic-parser -input "$batch" -output "results_$batch.json"
done

# Combine results
jq -s 'add' results_batch_*.json > all_results.json
```

## Architecture

```
comic-parser/
├── main.go                 # CLI entry point
├── config/
│   └── config.go          # Configuration management
├── llm/
│   └── client.go          # Anthropic API client
├── comicvine/
│   └── client.go          # ComicVine API client
├── models/
│   └── models.go          # Data structures
├── processor/
│   └── processor.go       # Main orchestration
└── prompts/
    └── prompts.go         # LLM prompt templates
```

## How It Works

1. **Filename Parsing**: Claude analyzes the filename and extracts structured data (title, issue number, year, etc.)

2. **ComicVine Search**: The app searches ComicVine's API using the extracted information

3. **Result Matching**: Claude reviews all ComicVine results and selects the best match based on title similarity, issue number, and date alignment

4. **Output**: Results are saved with full metadata and match confidence scores

## License

MIT
