# Project Data Raw Response — Design

**Date:** 2026-02-25
**Goal:** Expose the raw JSON from `GET /open/v1/project/{projectId}/data` so the actual response shape can be discovered before modeling columns.

## Context

The existing `ListTasks` client method calls this same endpoint but only decodes `{ tasks: [...] }`, silently discarding all other fields. The API may return additional data (columns, column assignments, etc.) that is currently invisible. This design adds a discovery command to reveal the full response.

## Architecture

### New client method

```go
// internal/client/projects.go
func (c *Client) GetProjectData(projectID string) (json.RawMessage, error)
```

Uses a new `doRawRequest(method, path string) ([]byte, error)` helper that returns the raw response body without struct decoding. Keeps existing `doRequest` untouched.

### New CLI command

```
dida365 project data <project-id>
```

Calls `GetProjectData`, pretty-prints the result via `json.Indent`, outputs to stdout. Lives in `cmd/project.go`.

## Data Flow

```
dida365 project data <project-id>
  → client.GetProjectData(projectID)
  → doRawRequest("GET", "/open/v1/project/{id}/data")
  → raw []byte
  → json.Indent (pretty print)
  → stdout
```

## Error Handling

- API errors (404, 401, etc.) → `outputError` with exit code 3
- JSON indent failure → error message to stderr

## Testing

- Unit test for `GetProjectData` using `httptest.NewServer`
- Verifies correct path is called and raw bytes returned unchanged
- No struct assertions — response shape is intentionally unknown

## What Comes Next

Once the raw response is inspected:
1. Model `column_id` on `Task` if present
2. Model a `Column` struct if a columns array is returned
3. Add column-aware task creation and move-task operations

## Files Changed

- Modify: `internal/client/projects.go` (add `doRawRequest`, `GetProjectData`)
- Modify: `cmd/project.go` (add `project data` subcommand)
- Modify: `internal/client/projects_test.go` (add test for `GetProjectData`)
