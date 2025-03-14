# Xander

Xander is a CLI application for retrieving metadata for media files from various online sources. It can parse filenames and fetch detailed information for comics and TV shows.

## Features

- **Comic Metadata**: Retrieve comic information from ComicVine API
- **TV Show Metadata**: Retrieve show information from TVDB (coming soon)
- **Flexible Parsing**: Works with various filename formats with or without extensions
- **Multiple Input Methods**: Process individual files or batch process from a text file
- **Configurable Output**: Choose between text and JSON output formats
- **API Rate Limiting**: Respects ComicVine's 200 requests/hour limit to prevent throttling
- **Response Caching**: Caches API responses for 24 hours to improve performance
- **Database Storage**: Save and query metadata in a local SQLite database
- **Export/Import**: Export database contents to JSON and import from JSON files

## Installation

### Prerequisites

- Go 1.23.5 or later

### Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd xander

# Build the binary
go build -o xander

# Optional: Move to a directory in your PATH
sudo mv xander /usr/local/bin/
```

## Configuration

Xander requires API keys for the services it interacts with. You can configure these keys using the CLI:

```bash
# Set ComicVine API key (required for comic metadata)
./xander config set-comicvine-key YOUR_COMICVINE_API_KEY

# Set TVDB API key (required for TV show metadata)
./xander config set-tvdb-key YOUR_TVDB_API_KEY

# View current configuration
./xander config
```

You can also set API keys using environment variables:

```bash
export XANDER_COMICVINE_API_KEY=your-api-key
export XANDER_TVDB_API_KEY=your-api-key
```

### Getting API Keys

- **ComicVine**: Register at [ComicVine](https://comicvine.gamespot.com/api/) to get an API key
- **TVDB**: Register at [TheTVDB](https://thetvdb.com/api-information) to get an API key

## Usage

### Comic Metadata

Get metadata for files with comic-like filenames using the ComicVine API (any file extension is supported):

```bash
# Process a single file
./xander comicvine "Batman (2016) #001.cbz"

# Process multiple files
./xander comicvine "Batman (2016) #001.cbz" "The Flash (2016) #001.cbr"

# Process files from a text file
./xander comicvine --input examples/comics.txt

# Output in JSON format
./xander comicvine --input examples/comics.txt --format json

# Enable verbose logging to debug API communication
./xander comicvine "Wonder Woman 013 (2024)" --verbose

# Store results in the local database
./xander comicvine --input examples/comics.txt --save
```

### Using Input Files

You can create a text file with a list of filenames to process (one per line):

```
# Example comics.txt
Batman (2016) #001.cbz
DC Comics - The Flash (2016) #001.cbr
Amazing Spider-Man Vol. 5 (2018) #001.cbz
```

### Supported Filename Formats

#### Comic Files

Supported comic filename formats (extension is optional):
- `Series (Year) #Issue`
- `Publisher - Series (Year) #Issue`
- `Series (Year) (digital) (Group)`
- `Series 001 (Year) (digital) (Group)`
- `Series - Title 000 (Year) (digital) (Group)`
- `Series v01 - Title (Year) (digital) (Group)`
- `Series 01 (of 08) (Year) (digital) (Group)`
- `YYYY-MM - Title (digital) (Group)`
- `YYYY (Year) (digital) (Group)`
- `Series.Title.Month.Year.Format.Group`
- `Series 001`

Examples:
- `Batman (2016) #001`
- `DC Comics - The Flash (2016) #001`
- `Amazing Spider-Man Vol. 5 (2018) #001`
- `Absolute Batman 001 (2024) (Webrip) (The Last Kryptonian-DCP).cbr`
- `Ultimate X-Men 007 (2024) (Digital) (Shan-Empire).cbz`
- `G.I. Joe - Cobra v01 (2009) (Minutemen-DarthTremens) (RC).cbz`
- `Jim Henson's Labyrinth 01 (of 08) (2024) (digital) (Son of Ultron-Empire).cbr`

Note: The parser ignores file extensions, so it can process files with any extension or no extension at all.

## API Rate Limiting

Xander respects the ComicVine API guidelines:
- 200 requests per hour limit
- Prevents velocity detection by adding small delays between requests
- Caches responses for 24 hours to reduce API calls
- Handles rate limit errors gracefully

When the `--verbose` flag is used, you'll see detailed information about:
- Rate limit status (current request count and reset time)
- Cache usage (hits, misses, and expiration times)
- API requests and responses

## Development

### Project Structure

```
xander/
├── cmd/             # CLI commands
├── internal/        # Internal packages
│   ├── comicvine/   # ComicVine API client
│   ├── config/      # Configuration handling
│   ├── parse/       # Filename parsing utilities
│   ├── storage/     # Database storage implementations
│   └── tvdb/        # TVDB API client (future)
├── examples/        # Example files
└── main.go          # Application entry point
```

### Testing

Run tests with:

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/parse

# Run tests with verbose output
go test -v ./...
```

### Code Style

This project follows Go standard code style and conventions. Ensure your code is properly formatted before submitting:

```bash
go fmt ./...
go vet ./...
```

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
