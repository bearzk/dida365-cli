# Dida365 CLI Design Document

**Date:** 2026-02-20
**Author:** Claude
**Status:** Approved

## Executive Summary

This document specifies a Go-based CLI for Dida365 (滴答清单), a task management application. The CLI targets automation and scripting workflows, providing full CRUD operations for tasks and read-only access to projects. It uses OAuth authentication with manual token configuration, outputs JSON exclusively for machine parsing, and follows a clean architecture pattern using Cobra and Viper.

## Requirements

### Primary Use Case
Automation and scripting integration with Dida365, enabling:
- Task creation from external systems (webhooks, monitoring tools, CI/CD)
- Task queries to trigger actions based on status or content
- Task updates, completion, and deletion in automated workflows
- Project listing and inspection for context

### Scope (MVP)
- **Tasks**: Create, read, update, delete, complete, list
- **Projects**: List, get details (read-only)
- **Authentication**: Manual token configuration via config file
- **Output**: JSON only (no tables or human-friendly formatting)
- **Error handling**: Structured JSON errors with exit codes

### Out of Scope (Future)
- Subtasks, priorities, due dates, reminders, recurring tasks
- Interactive OAuth flow with browser
- Project creation, update, deletion
- Filtering, sorting, batch operations
- Multiple output formats (YAML, tables, templates)

## Architecture

### Technology Stack
- **Language**: Go
- **CLI Framework**: Cobra (command structure, flag parsing, help generation)
- **Config Management**: Viper (read/write `~/.dida365/config.json`)
- **HTTP Client**: Standard library `net/http`
- **JSON Processing**: Standard library `encoding/json`

### Project Structure
```
dida365-cli/
├── cmd/
│   ├── root.go           # Root command, version, global flags
│   ├── auth.go           # auth configure, auth status
│   ├── task.go           # task create/get/list/update/complete/delete
│   └── project.go        # project list/get
├── internal/
│   ├── client/
│   │   ├── client.go     # HTTP client, OAuth headers, error parsing
│   │   ├── tasks.go      # Task API operations
│   │   └── projects.go   # Project API operations
│   ├── config/
│   │   └── config.go     # Load/save/validate config file
│   └── models/
│       ├── task.go       # Task structs, JSON tags
│       └── project.go    # Project structs, JSON tags
├── main.go               # Entry point
├── go.mod
├── go.sum
└── README.md
```

### Command Structure
```
dida365 auth configure          # Set client_id, client_secret, access_token
dida365 auth status             # Validate config and token

dida365 task create             # --title (required), --project-id (required), --content (optional)
dida365 task get <task-id>      # --project-id (required)
dida365 task list <project-id>  # List all tasks in project
dida365 task update <task-id>   # --project-id (required), --title, --content
dida365 task complete <task-id> # --project-id (required)
dida365 task delete <task-id>   # --project-id (required)

dida365 project list            # List all projects
dida365 project get <project-id> # Get project details
```

## API Client Design

### Client Interface
The `internal/client` package provides a reusable API client:

```go
type Client struct {
    httpClient *http.Client
    config     *config.Config
    baseURL    string
}

func NewClient(cfg *config.Config) *Client
func (c *Client) doRequest(method, path string, body, result interface{}) error
```

The `doRequest` method handles:
- Authorization header injection (`Authorization: Bearer {token}`)
- JSON marshaling for request bodies
- JSON unmarshaling for responses
- HTTP error translation to Go errors with context

### API Operations

**tasks.go:**
```go
func (c *Client) CreateTask(task *models.TaskCreate) (*models.Task, error)
func (c *Client) GetTask(projectId, taskId string) (*models.Task, error)
func (c *Client) ListTasks(projectId string) ([]*models.Task, error)
func (c *Client) UpdateTask(projectId, taskId string, updates *models.TaskUpdate) (*models.Task, error)
func (c *Client) CompleteTask(projectId, taskId string) error
func (c *Client) DeleteTask(projectId, taskId string) error
```

**projects.go:**
```go
func (c *Client) ListProjects() ([]*models.Project, error)
func (c *Client) GetProject(projectId string) (*models.Project, error)
```

### Error Mapping
- **401 Unauthorized** → "access token expired or invalid"
- **403 Forbidden** → "insufficient permissions for this operation"
- **404 Not Found** → "resource not found: {type} {id}"
- **400 Bad Request** → Parse API error message from response body
- **500 Server Error** → "Dida365 server error, try again later"
- **Network errors** → Wrap with context (timeout, connection refused, DNS)

## Data Models

### Task Model
```go
type Task struct {
    ID            string     `json:"id"`
    ProjectID     string     `json:"projectId"`
    Title         string     `json:"title"`
    Content       string     `json:"content,omitempty"`
    Status        int        `json:"status"`           // 0=normal, 2=completed
    Priority      int        `json:"priority,omitempty"` // 0=none, 1=low, 3=med, 5=high
    CompletedTime *time.Time `json:"completedTime,omitempty"`
    SortOrder     int        `json:"sortOrder"`
}

type TaskCreate struct {
    Title     string `json:"title"`
    ProjectID string `json:"projectId"`
    Content   string `json:"content,omitempty"`
}

type TaskUpdate struct {
    Title   *string `json:"title,omitempty"`
    Content *string `json:"content,omitempty"`
}
```

### Project Model
```go
type Project struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Color     string `json:"color,omitempty"`
    SortOrder int    `json:"sortOrder"`
    Closed    bool   `json:"closed"`
    Kind      string `json:"kind"` // "TASK" or "NOTE"
}
```

### JSON Output
All commands output JSON to stdout:

**Success:**
```bash
$ dida365 task create --title "Deploy v2" --project-id "proj123"
{"id":"task456","projectId":"proj123","title":"Deploy v2","status":0,"sortOrder":0}

$ dida365 project list
[{"id":"proj1","name":"Personal","closed":false,"kind":"TASK"},{"id":"proj2","name":"Work","closed":false,"kind":"TASK"}]
```

**Errors** (stderr):
```bash
$ dida365 task get invalid-id --project-id "proj123"
{"error":"task not found: task invalid-id","code":"NOT_FOUND"}
```

## Configuration Management

### Config File
Location: `~/.dida365/config.json`
Permissions: `0600` (user read/write only)

```json
{
  "client_id": "your_client_id",
  "client_secret": "your_client_secret",
  "access_token": "your_access_token",
  "base_url": "https://dida365.com"
}
```

### Configuration Flow

**Initial setup:**
```bash
$ dida365 auth configure \
    --client-id "abc123" \
    --client-secret "secret456" \
    --access-token "token789"
{"configured":true,"config_path":"/Users/user/.dida365/config.json"}
```

The command:
1. Creates `~/.dida365/` with `0700` permissions
2. Writes config file with `0600` permissions
3. Validates token with test API call (`GET /open/v1/project`)
4. Returns success JSON or error

**Status check:**
```bash
$ dida365 auth status
{"configured":true,"client_id":"abc123","token_valid":true}
```

### Config Behavior
- All commands (except `auth configure`) load config automatically
- Missing or invalid config → JSON error, exit code 1
- Commands validate config before API calls
- Config validation checks required fields and JSON syntax

## Error Handling

### Exit Codes
- **0**: Success
- **1**: Configuration error (missing, invalid, or unreadable config)
- **2**: Authentication error (401, 403)
- **3**: API error (400, 404, 500)
- **4**: Network error (timeout, connection refused, DNS)
- **5**: Client error (invalid arguments, missing flags)

### Error Output Format
All errors output to stderr as JSON:
```json
{
  "error": "human-readable message",
  "code": "ERROR_CODE",
  "details": {}  // optional
}
```

### Examples
```bash
# Missing flag
$ dida365 task create --title "Test"
{"error":"missing required flag: --project-id","code":"MISSING_FLAG"}
# exit 5

# Expired token
$ dida365 task list proj123
{"error":"access token expired or invalid","code":"UNAUTHORIZED"}
# exit 2

# Not found
$ dida365 task get bad-id --project-id "proj123"
{"error":"task not found: task bad-id","code":"NOT_FOUND"}
# exit 3
```

### Scripting Integration
Success outputs go to stdout for piping:
```bash
dida365 task create --title "Deploy" --project-id "p1" | jq .id
dida365 project list | jq '.[0].name'
dida365 task list proj123 | jq 'map(select(.status == 0))'
```

## Testing Strategy

### Test Coverage

**Unit Tests (Priority):**
- `internal/client/` - Mock HTTP responses with `httptest.NewServer()`
  - Test all HTTP methods (GET, POST, DELETE)
  - Test error conditions (401, 404, 500, network errors)
  - Verify JSON marshaling/unmarshaling
  - Test authorization header injection
- `internal/config/` - Use temporary directories for config I/O
  - Test load/save with valid/invalid JSON
  - Test missing config file handling
  - Test permission errors
  - Test field validation
- `internal/models/` - JSON marshaling tests
  - Verify struct tags match API format
  - Test omitempty behavior
  - Test time.Time serialization

**Integration Tests (Secondary):**
- End-to-end command tests with mock API server
- Config flow: configure → status → task operations
- Error propagation from client → commands → output

### Test Structure
```
internal/
├── client/
│   ├── client_test.go
│   ├── tasks_test.go
│   └── projects_test.go
├── config/
│   └── config_test.go
└── models/
    ├── task_test.go
    └── project_test.go
```

### Testing Approach
- Use table-driven tests for multiple scenarios
- Mock HTTP with `httptest.NewServer()` for predictable responses
- Test happy paths and all error branches
- Verify JSON output format matches expected structure
- Use `t.TempDir()` for config file tests

## Implementation Notes

### Dida365 API Reference
- **Base URL**: `https://dida365.com`
- **Auth**: OAuth2 Bearer token in `Authorization` header
- **Content-Type**: `application/json`
- **Task endpoints**: `/open/v1/task`, `/open/v1/project/{projectId}/task/{taskId}`
- **Project endpoints**: `/open/v1/project`, `/open/v1/project/{projectId}`

### Authentication Setup
Users must manually obtain credentials from Dida365 developer portal:
1. Register application at developer.dida365.com
2. Note `client_id` and `client_secret`
3. Complete OAuth flow externally to obtain `access_token`
4. Configure CLI with `dida365 auth configure`

Future versions may add interactive OAuth flow with browser launch and callback handling.

### Extensibility
The architecture supports future additions:
- Additional task fields (due dates, priorities, subtasks, reminders)
- Project CRUD operations
- Filtering and sorting flags
- Batch operations
- Alternative output formats (use `--format` flag)
- Plugin system for custom commands

## Success Criteria

The CLI is successful if:
1. **Scripting-friendly**: JSON output parsable with `jq` or standard JSON tools
2. **Reliable**: Clear error messages with appropriate exit codes
3. **Maintainable**: Clean architecture enables adding features without refactoring
4. **Well-tested**: Core client library has comprehensive unit tests
5. **Documented**: README with setup instructions, command examples, scripting patterns

## Conclusion

This design provides a solid foundation for a Dida365 CLI focused on automation workflows. The Cobra framework, clean architecture, and JSON-only output ensure the tool integrates seamlessly into scripts and CI/CD pipelines. The MVP scope keeps initial development focused while the architecture remains extensible for future enhancements.
