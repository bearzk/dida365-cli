# Column Support Design

**Date:** 2026-02-25

## Context

The Dida365 API `GET /open/v1/project/{projectId}/data` returns a typed response with `project`, `tasks`, and `columns`. Tasks carry a `columnId` field. Moving a task between columns is a `POST /open/v1/task/{taskId}` with `columnId` in the body — the existing `UpdateTask` endpoint.

## What We're Building

Five changes, each self-contained:

1. **`Column` model** — new struct in `internal/models/`
2. **`columnId` on `Task`** — add field to existing model
3. **`columnId` on `TaskUpdate`** — allow column moves via UpdateTask
4. **`ProjectData` model** — typed struct replacing raw `[]byte`
5. **`GetProjectData` typed** — returns `*models.ProjectData`, removes `doRawRequest`
6. **`project columns` command** — list columns for a project
7. **`task move` command** — move task to a column by ID

## Data Model

```go
// Column represents a Kanban column within a project
type Column struct {
    ID        string `json:"id"`
    ProjectID string `json:"projectId"`
    Name      string `json:"name"`
    SortOrder int64  `json:"sortOrder"`
}

// ProjectData is the full response from GET /open/v1/project/{id}/data
type ProjectData struct {
    Project Project  `json:"project"`
    Tasks   []*Task  `json:"tasks"`
    Columns []Column `json:"columns"`
}

// Task gains one new field:
ColumnID string `json:"columnId,omitempty"`

// TaskUpdate gains one new field:
ColumnID *string `json:"columnId,omitempty"`
```

`sortOrder` is `int64` — the API returns values like `-4611686018426863616` which overflow `int32`.

## Data Flow

```
dida365 project columns <project-id>
  → client.GetProjectData(projectID) → *models.ProjectData
  → outputJSON(data.Columns)

dida365 task move <task-id> --column-id <col-id>
  → client.UpdateTask(taskID, &TaskUpdate{ColumnID: &colID})
  → POST /open/v1/task/{taskId} {"columnId": "..."}
  → outputJSON(updatedTask)

dida365 project data <project-id>
  → client.GetProjectData(projectID) → *models.ProjectData
  → outputJSON(data)   ← now typed, not raw bytes
```

## Client Changes

- `GetProjectData(projectID string) (*models.ProjectData, error)` — replaces raw `[]byte` version, uses `doRequest` with the typed struct
- `doRawRequest` — **removed** (discovery job done)

## Commands

### `dida365 project columns <project-id>`
- Args: ExactArgs(1)
- Calls `GetProjectData`, returns `data.Columns`

### `dida365 task move <task-id> --column-id <col-id>`
- Args: ExactArgs(1)
- Flag: `--column-id` (required)
- Calls `UpdateTask` with only `ColumnID` set

## Error Handling

All errors follow existing pattern: `outputError(err, "API_ERROR", 3)`.

## Testing

- `TestColumn` — JSON marshal/unmarshal, int64 sortOrder
- `TestProjectData` — marshal/unmarshal round-trip
- `TestGetProjectData` — mock server returning typed response
- Build + help output for new commands
