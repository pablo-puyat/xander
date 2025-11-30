## Project Overview

This is a Go application that parses comic book archive filenames (CBR/CBZ) and matches them against ComicVine's database using LLM-powered parsing and selection.

**Problem it solves**: Comic filenames come in wildly inconsistent formats. Traditional regex parsers fail on edge cases. This app uses Claude to intelligently parse filenames and select the best match from ComicVine's often-ambiguous search results.

## Quick Reference

```bash
# Build
go build -o comic-parser ./cmd/comic-parser

# Test single file
./comic-parser -file "Amazing Spider-Man 001 (2018).cbz" -verbose

# Batch process
./comic-parser -input filenames.txt -output results.json

# Generate sample config
./comic-parser -generate-config
```

## Architecture

```
comic-parser/
├── cmd/
│   └── comic-parser/
│       └── main.go         # CLI entry point, flag parsing, output handling
├── internal/
│   ├── config/config.go        # Configuration from env vars and JSON file
│   ├── llm/client.go           # Anthropic API client (Claude)
│   ├── comicvine/client.go     # ComicVine API client with rate limiting
│   ├── models/models.go        # All data structures
│   ├── processor/processor.go  # Main orchestration, worker pool
│   └── prompts/prompts.go      # LLM prompt templates (CRITICAL)
```

## Code Organization & Go Best Practices

This codebase follows idiomatic Go patterns and conventions:

### Package Structure
- **Package documentation**: Every package has a doc comment explaining its purpose
- **Exported identifiers**: All public functions, types, and constants are documented
- **Package naming**: Short, lowercase, single-word package names without underscores

### Constants and Configuration
- **Magic numbers eliminated**: All hardcoded values extracted to named constants
- **Grouped constants**: Related constants grouped with descriptive comments
- **Examples**:
  - `config/config.go`: Default values as constants with `default` prefix
  - `llm/client.go`: API configuration constants (headers, timeouts, versions)
  - `comicvine/client.go`: API parameters, limits, and format strings as constants

### Error Handling
- **Error wrapping**: All errors wrapped with `fmt.Errorf(...: %w, err)` for context
- **Graceful degradation**: Processing errors captured in results, batch processing continues
- **Context cancellation**: All operations respect `context.Context` for graceful shutdown

### Concurrency Patterns
- **Thread-safe access**: Mutexes protect shared state (caches, progress tracking)
- **RWMutex for caches**: Read-write locks optimize concurrent read access
- **Worker pools**: Controlled concurrency with configurable worker count
- **Channel communication**: Results passed via channels for coordination

### Type Safety and Clarity
- **Structured types**: All data modeled with explicit structs, not maps
- **Field ordering**: Exported fields before unexported, logical grouping
- **JSON tags**: Proper snake_case JSON serialization tags
- **Pointer usage**: Pointers used for optional/nullable fields (`*ComicVineIssue`)

### API Client Patterns
- **Rate limiting**: Built-in rate limiting using `time.Ticker`
- **Caching**: Volume cache to reduce redundant API calls
- **Retries**: Configurable retry logic with exponential backoff
- **Timeouts**: HTTP clients configured with reasonable timeouts
- **Context propagation**: All requests accept and respect context

## Key Design Decisions

### Two-Stage LLM Approach
1. **Parse prompt**: Extracts title, issue number, year from filename
2. **Match prompt**: Selects best result from ComicVine candidates

This separation allows each prompt to be optimized independently.

### ComicVine Search Strategy
The app searches volumes first, then issues within matching volumes. This is more reliable than ComicVine's general search. Falls back to direct issue search if volume search fails.

### Rate Limiting
- ComicVine: ~1 request/second (built into client)
- Anthropic: Configurable via `rate_limit_per_min`
- Worker count controls parallelism

## Working with the Codebase

### Code Quality Tools
Before committing changes, always run:
```bash
go fmt ./...      # Format all code to Go standards
go vet ./...      # Static analysis for common errors
go build ./...    # Ensure everything compiles
```

### Adding New Filename Patterns
Edit `prompts/prompts.go` → `FilenameParsePrompt()`. Add examples to the prompt showing the new pattern. The LLM learns from examples.

### Improving Match Accuracy
Edit `prompts/prompts.go` → `ResultMatchPrompt()`. Adjust the matching rules or add edge cases to the prompt.

### Adding New Output Formats
Edit `main.go` → `saveResults()`. Add a new case in the switch statement. Follow the pattern of `saveJSON()` and `saveCSV()`.

### Changing ComicVine Search Behavior
Edit `comicvine/client.go` → `SearchIssues()` or `searchByVolumeAndIssue()`. The search strategy is here.

### Adding New Constants
When adding hardcoded values:
1. Define as a constant at package level with descriptive name
2. Group with related constants
3. Add a comment explaining the purpose or source
4. Use the constant throughout the code instead of the literal value

## API Details

### Anthropic API
- Endpoint: `POST /v1/messages`
- Model: `claude-sonnet-4-20250514` (configurable)
- Auth: `x-api-key` header
- Version header required: `anthropic-version: 2023-06-01`

### ComicVine API
- Base URL: `https://comicvine.gamespot.com/api`
- Auth: `api_key` query parameter
- Always set `format=json`
- User-Agent header recommended
- Rate limit: Be conservative, ~1 req/sec

## Important Patterns

### JSON Extraction from LLM
LLMs sometimes wrap JSON in markdown. The `llm.ExtractJSON()` function handles this:
```go
jsonStr := llm.ExtractJSON(response)  // Strips ```json blocks, finds {}
```

### Error Handling
Processing errors are captured in results, not thrown. This allows batch processing to continue:
```go
result.Error = fmt.Sprintf("parsing filename: %v", err)
return result, nil  // Return result with error, not error
```

### Context Cancellation
All operations respect `context.Context` for graceful shutdown:
```go
select {
case <-ctx.Done():
    return
default:
}
```

### Thread Safety
When accessing shared state from multiple goroutines:
```go
// Use sync.Mutex for exclusive access
p.progressMu.Lock()
p.progress.Processed++
p.progressMu.Unlock()

// Use sync.RWMutex for read-heavy caches
c.cacheMutex.RLock()
if vol, ok := c.volumeCache[volumeID]; ok {
    c.cacheMutex.RUnlock()
    return vol, nil
}
c.cacheMutex.RUnlock()
```

### Constant Usage
Use named constants instead of magic strings/numbers:
```go
// Good
params.Set(paramAPIKey, c.apiKey)
httpReq.Header.Set(headerVersion, anthropicVersion)

// Bad - avoid
params.Set("api_key", c.apiKey)
httpReq.Header.Set("anthropic-version", "2023-06-01")
```

## Common Tasks

### "Add support for a new comic naming convention"
1. Add example to `FilenameParsePrompt()` in `prompts/prompts.go`
2. Test with: `./comic-parser -file "example filename.cbz" -verbose`

### "Improve matching when multiple volumes have same name"
1. Edit matching rules in `ResultMatchPrompt()` in `prompts/prompts.go`
2. Consider adding more fields to `SimpleResult` struct in the prompt

### "Add a new field to track (e.g., variant covers)"
1. Add field to `ParsedFilename` in `models/models.go`
2. Update `FilenameParsePrompt()` to extract it
3. Update output functions in `main.go` if needed

### "Cache ComicVine results to reduce API calls"
Volume cache already exists in `comicvine/client.go`. Extend the pattern:
```go
type Client struct {
    volumeCache map[int]*models.ComicVineVolume
    issueCache  map[string][]models.ComicVineIssue  // Add this
}
```

### "Add resume capability for interrupted batches"
1. Check if output file exists at startup
2. Load existing results and extract processed filenames
3. Filter input list to skip already-processed files

## Testing

No test files currently exist. When adding tests:

```go
// processor/processor_test.go
func TestParseFilename(t *testing.T) {
    // Mock the LLM client or use recorded responses
}
```

For integration testing, create a small test file list and use real APIs with `-verbose`.

## Environment Variables

```bash
ANTHROPIC_API_KEY    # Required - Anthropic API key
COMICVINE_API_KEY    # Required - ComicVine API key
```

## Config File Fields

```json
{
  "anthropic_api_key": "",           // Can also use env var
  "comicvine_api_key": "",           // Can also use env var
  "anthropic_model": "claude-sonnet-4-20250514",
  "anthropic_max_tokens": 1024,
  "worker_count": 3,                 // Concurrent processors
  "rate_limit_per_min": 30,          // Anthropic rate limit
  "retry_attempts": 3,
  "retry_delay_seconds": 2,
  "verbose": false
}
```

## Gotchas

1. **ComicVine issue numbers are strings** - "1", "1.1", "Annual 1" are all valid
2. **Volume IDs use format 4050-{id}** in some endpoints
3. **LLM responses may have trailing text** - Always use `ExtractJSON()`
4. **ComicVine search is fuzzy** - "Spider-Man" matches "Spider-Man 2099"
5. **Cover dates vs store dates** - Cover dates are often 2-3 months ahead

## Performance Notes

- Each file requires 2 LLM API calls + 2-5 ComicVine API calls
- Expect ~10-30 seconds per file depending on API latency
- For 11,000 files: estimate 30-90 hours with 3 workers
- Consider batching into chunks of 500-1000 for manageability

## Future Improvements to Consider

- [ ] Persistent cache (SQLite or file-based)
- [ ] Resume interrupted batch processing
- [ ] Parallel ComicVine searches per file
- [ ] Alternative metadata sources (GCD, CLZ, etc.)
- [ ] Local LLM support (Ollama)
- [ ] Web UI for reviewing/correcting matches
- [ ] Export to comic management software formats
