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

### Authentication Flow
✅ **Completed**: Full OAuth2 flow implemented

**Commands:**
- `auth login` - Complete OAuth2 flow (login + token exchange)
- `auth status` - Check authentication status

**Flow:**
1. Local callback server starts on configurable port (default: 8080)
2. Browser opens to authorization URL (dida365.com or ticktick.com)
3. User authorizes, redirected to `http://localhost:PORT/callback`
4. CLI exchanges auth code for access_token
5. Token saved to `~/.dida365/config.json` with expiry time

**Token Management:**
- Access tokens are long-lived (~6 months)
- The Dida365 API does NOT issue refresh tokens
- Re-authenticate with `auth login` when token expires

**Services Supported:**
- `dida365` - China service (dida365.com)
- `ticktick` - International service (ticktick.com)

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
