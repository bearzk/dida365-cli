# Dida365 CLI - Project Context

## Quick Reference

- **Module**: `github.com/bearzk/dida365-cli`
- **Config**: `~/.dida365/config.json` (0600 permissions)
- **Tests**: `go test ./... -v` (must pass before commits)
- **Build**: `go build -o dida365 .`

## Code Patterns

- JSON fields: snake_case (`client_id`, not `clientId`)
- Error wrapping: `fmt.Errorf("context: %w", err)` for error chains
- Exit codes: 0=success, 1=config, 2=auth, 3=API, 5=validation
- CLI output: All commands return JSON (automation-friendly)

## Critical Implementation Details

### UpdateTask Signature
- Correct: `UpdateTask(taskID string, updates *models.TaskUpdate)`
- Takes only taskID, NOT projectID (fixed in Task 7)

### Authentication Flow (TODO)
⚠️ **Current limitation**: Requires manual `access_token` input

**Should implement OAuth2 flow:**
1. Start local server on `http://localhost:8080/callback`
2. Open browser to authorization URL
3. Exchange auth code for access_token
4. Save token with refresh_token

Reference: https://cyfine.github.io/TickTick-Dida365-API-Client/guides/authentication/

## Testing Approach

- TDD enforced: Write tests first, implementation second
- Mock API calls with `httptest.NewServer`
- 100% test pass rate required before proceeding
- Current coverage: 27 tests across 3 packages

## Project Structure

```
cmd/              # CLI commands (auth, project, task)
internal/
  ├── client/     # API client (8 operations)
  ├── config/     # Config management
  └── models/     # Data models
main.go           # Entry point
```

## User Preferences

- Use DuckDuckGo search skill (not WebSearch tool) for web searches
