# Xander Project Guidelines

## Build Commands
```bash
# Build the project
go build -o xander

# Run the application
go run main.go

# Run tests
go test ./...

# Run a specific test
go test ./path/to/package -run TestName

# Format code
go fmt ./...

# Lint code
go vet ./...

# Run the comic processor
./xander comic --input examples/comics.txt
```

## Project Structure
```
xander/
├── cmd/             # CLI commands
│   ├── config.go    # Configuration command
│   ├── comic.go     # Comic processing command
│   └── root.go      # Root CLI command
├── internal/        # Internal packages
│   ├── comicvine/   # ComicVine API client
│   ├── config/      # Configuration handling
│   ├── parse/       # Filename parsing utilities
│   └── tvdb/        # TVDB API client
├── examples/        # Example files
└── main.go          # Application entry point
```

## Code Style Guidelines
- **Imports**: Standard library first, third-party packages second, internal packages last
- **Naming**: Use CamelCase for exported names, camelCase for internal names
- **Error Handling**: Use fmt.Errorf with %w for wrapping errors with context
- **Comments**: Follow GoDoc style with function descriptions
- **Packages**: Use lowercase, single-word names
- **CLI Structure**: Use cobra for command management with init() pattern
- **Config**: Prefer environment variables with config file fallbacks
- **Logging**: Use standard log package with appropriate output control

## Testing Approach
- Follow TDD (Test-Driven Development) principles
- Write tests before implementing features
- Use table-driven tests for comprehensive test cases
- Mock external dependencies like API clients
- Test both success and error paths

## Configuration Management
- Configuration keys:
  - XANDER_API_KEY: General application API key
  - XANDER_COMICVINE_API_KEY: ComicVine API key
  - XANDER_TVDB_API_KEY: TVDB API key
- Config file location: ~/.config/xander/config.yaml

## Feature Roadmap
- Comic metadata retrieval from ComicVine (implemented)
- TV show metadata retrieval from TVDB (planned)
- Movie metadata retrieval (future)
- Local database for caching results (future)
- GUI interface (future)