# Column Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Add Column model, expose column_id on Task/TaskUpdate, replace raw GetProjectData with a typed version, and add `project columns` and `task move` CLI commands.

**Architecture:** Four tasks in dependency order — models first, then client, then commands. Moving a task between columns reuses the existing UpdateTask endpoint with a new ColumnID field on TaskUpdate. No new HTTP infrastructure needed.

**Tech Stack:** Go stdlib, Cobra CLI, existing `doRequest` client infrastructure.

---

## Task 1: Add Column model and update Task/TaskUpdate models

**Files:**
- Create: `internal/models/column.go`
- Create: `internal/models/column_test.go`
- Modify: `internal/models/task.go`
- Modify: `internal/models/task_test.go`

**Context:** The Dida365 API returns `sortOrder` values as large negative int64s (e.g. `-4611686018426863616`). Use `int64` not `int`. The existing `Task` model is missing `columnId`. `TaskUpdate` uses pointer fields with `omitempty` — follow the same pattern for `ColumnID`.

---

**Step 1: Write failing tests for Column model**

Create `internal/models/column_test.go`:

```go
package models

import (
	"encoding/json"
	"testing"
)

func TestColumnJSONMarshaling(t *testing.T) {
	t.Run("marshal column", func(t *testing.T) {
		col := Column{
			ID:        "6998df0d7fff114cc5fb5afc",
			ProjectID: "6998dee97fe9514cc5fb5afa",
			Name:      "Backlog",
			SortOrder: -4611686018426863616,
		}

		data, err := json.Marshal(col)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}

		var got Column
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if got.ID != col.ID {
			t.Errorf("ID: got %s, want %s", got.ID, col.ID)
		}
		if got.Name != col.Name {
			t.Errorf("Name: got %s, want %s", got.Name, col.Name)
		}
		if got.SortOrder != col.SortOrder {
			t.Errorf("SortOrder: got %d, want %d", got.SortOrder, col.SortOrder)
		}
	})

	t.Run("unmarshal from api response", func(t *testing.T) {
		raw := `{
			"id": "6998df0d7fff114cc5fb5afc",
			"projectId": "6998dee97fe9514cc5fb5afa",
			"name": "Backlog",
			"sortOrder": -4611686018426863616
		}`

		var col Column
		if err := json.Unmarshal([]byte(raw), &col); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if col.ID != "6998df0d7fff114cc5fb5afc" {
			t.Errorf("ID: got %s", col.ID)
		}
		if col.SortOrder != -4611686018426863616 {
			t.Errorf("SortOrder: got %d, want -4611686018426863616", col.SortOrder)
		}
	})
}

func TestProjectDataJSONMarshaling(t *testing.T) {
	t.Run("unmarshal full project data response", func(t *testing.T) {
		raw := `{
			"project": {"id": "proj1", "name": "Test", "sortOrder": 0, "closed": false, "kind": "TASK"},
			"tasks": [
				{"id": "task1", "projectId": "proj1", "title": "A task", "status": 0, "sortOrder": 0, "columnId": "col1"}
			],
			"columns": [
				{"id": "col1", "projectId": "proj1", "name": "Backlog", "sortOrder": -100}
			]
		}`

		var pd ProjectData
		if err := json.Unmarshal([]byte(raw), &pd); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if pd.Project.ID != "proj1" {
			t.Errorf("Project.ID: got %s", pd.Project.ID)
		}
		if len(pd.Tasks) != 1 {
			t.Fatalf("Tasks: got %d, want 1", len(pd.Tasks))
		}
		if pd.Tasks[0].ColumnID != "col1" {
			t.Errorf("Tasks[0].ColumnID: got %s, want col1", pd.Tasks[0].ColumnID)
		}
		if len(pd.Columns) != 1 {
			t.Fatalf("Columns: got %d, want 1", len(pd.Columns))
		}
		if pd.Columns[0].Name != "Backlog" {
			t.Errorf("Columns[0].Name: got %s, want Backlog", pd.Columns[0].Name)
		}
	})
}
```

**Step 2: Run to verify failure**

```bash
go test ./internal/models -v -run "TestColumnJSONMarshaling|TestProjectDataJSONMarshaling"
```

Expected: FAIL with `undefined: Column`, `undefined: ProjectData`

**Step 3: Create `internal/models/column.go`**

```go
package models

// Column represents a Kanban column within a project.
type Column struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	SortOrder int64  `json:"sortOrder"`
}

// ProjectData is the full typed response from GET /open/v1/project/{id}/data.
type ProjectData struct {
	Project Project  `json:"project"`
	Tasks   []*Task  `json:"tasks"`
	Columns []Column `json:"columns"`
}
```

**Step 4: Add `ColumnID` to `Task` in `internal/models/task.go`**

Add one field to the `Task` struct after `SortOrder`:

```go
ColumnID string `json:"columnId,omitempty"`
```

The full struct becomes:

```go
type Task struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"projectId"`
	Title         string     `json:"title"`
	Content       string     `json:"content,omitempty"`
	Status        int        `json:"status"`
	Priority      int        `json:"priority,omitempty"`
	CompletedTime *time.Time `json:"completedTime,omitempty"`
	SortOrder     int        `json:"sortOrder"`
	ColumnID      string     `json:"columnId,omitempty"`
}
```

**Step 5: Add `ColumnID` to `TaskUpdate` in `internal/models/task.go`**

Add one field to the `TaskUpdate` struct:

```go
ColumnID *string `json:"columnId,omitempty"`
```

The full struct becomes:

```go
type TaskUpdate struct {
	Title    *string `json:"title,omitempty"`
	Content  *string `json:"content,omitempty"`
	ColumnID *string `json:"columnId,omitempty"`
}
```

**Step 6: Add a TaskUpdate test for ColumnID in `internal/models/task_test.go`**

Add a sub-test inside `TestTaskUpdateJSONMarshaling`:

```go
t.Run("set column id", func(t *testing.T) {
	colID := "6998df0d7fff114cc5fb5afc"
	update := TaskUpdate{
		ColumnID: &colID,
	}

	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got["columnId"] != colID {
		t.Errorf("columnId: got %v, want %s", got["columnId"], colID)
	}
	if _, ok := got["title"]; ok {
		t.Error("title should be omitted when nil")
	}
})
```

**Step 7: Run all model tests**

```bash
go test ./internal/models -v
```

Expected: All PASS

**Step 8: Commit**

```bash
git add internal/models/column.go internal/models/column_test.go internal/models/task.go internal/models/task_test.go
git commit -m "feat(models): add Column, ProjectData, and columnId fields on Task and TaskUpdate"
```

---

## Task 2: Update GetProjectData to return typed response, remove doRawRequest

**Files:**
- Modify: `internal/client/client.go` — remove `doRawRequest`
- Modify: `internal/client/projects.go` — update `GetProjectData` signature
- Modify: `internal/client/projects_test.go` — update `TestGetProjectData`
- Modify: `cmd/project.go` — update `runProjectData` handler

**Context:** `GetProjectData` currently returns `([]byte, error)` via `doRawRequest`. We replace it to return `(*models.ProjectData, error)` via the existing `doRequest`. The `project data` command then uses `outputJSON` like every other command — no `json.Indent`, no `bytes.Buffer`.

---

**Step 1: Update `TestGetProjectData` in `internal/client/projects_test.go`**

Replace the entire `TestGetProjectData` function with:

```go
func TestGetProjectData(t *testing.T) {
	t.Run("returns typed project data", func(t *testing.T) {
		projectData := map[string]interface{}{
			"project": map[string]interface{}{
				"id": "proj123", "name": "Test", "sortOrder": 0, "closed": false, "kind": "TASK",
			},
			"tasks": []map[string]interface{}{
				{"id": "t1", "projectId": "proj123", "title": "A task", "status": 0, "sortOrder": 0, "columnId": "col1"},
			},
			"columns": []map[string]interface{}{
				{"id": "col1", "projectId": "proj123", "name": "Backlog", "sortOrder": -100},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("method: got %s, want GET", r.Method)
			}
			if r.URL.Path != "/open/v1/project/proj123/data" {
				t.Errorf("path: got %s, want /open/v1/project/proj123/data", r.URL.Path)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer test" {
				t.Errorf("Authorization: got %q, want %q", got, "Bearer test")
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(projectData)
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID: "test", ClientSecret: "test", AccessToken: "test", BaseURL: server.URL,
		}

		c := NewClient(cfg)
		result, err := c.GetProjectData("proj123")
		if err != nil {
			t.Fatalf("GetProjectData failed: %v", err)
		}

		if result.Project.ID != "proj123" {
			t.Errorf("Project.ID: got %s, want proj123", result.Project.ID)
		}
		if len(result.Tasks) != 1 {
			t.Fatalf("Tasks: got %d, want 1", len(result.Tasks))
		}
		if result.Tasks[0].ColumnID != "col1" {
			t.Errorf("Tasks[0].ColumnID: got %s, want col1", result.Tasks[0].ColumnID)
		}
		if len(result.Columns) != 1 {
			t.Fatalf("Columns: got %d, want 1", len(result.Columns))
		}
		if result.Columns[0].Name != "Backlog" {
			t.Errorf("Columns[0].Name: got %s, want Backlog", result.Columns[0].Name)
		}
	})

	t.Run("returns error on api failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":"not found"}`)
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID: "test", ClientSecret: "test", AccessToken: "test", BaseURL: server.URL,
		}

		c := NewClient(cfg)
		_, err := c.GetProjectData("missing")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})
}
```

**Step 2: Run to verify failure**

```bash
go test ./internal/client -v -run TestGetProjectData
```

Expected: FAIL — compile error because `GetProjectData` still returns `[]byte`

**Step 3: Update `GetProjectData` in `internal/client/projects.go`**

Replace the current `GetProjectData` function with:

```go
// GetProjectData returns the full typed response from the project data endpoint,
// including the project details, tasks, and columns.
// For normal task listing use ListTasks instead.
func (c *Client) GetProjectData(projectID string) (*models.ProjectData, error) {
	path := fmt.Sprintf("/open/v1/project/%s/data", projectID)
	var data models.ProjectData
	if err := c.doRequest("GET", path, nil, &data); err != nil {
		return nil, fmt.Errorf("failed to get project data: %w", err)
	}
	return &data, nil
}
```

Make sure the import block includes `"github.com/bearzk/dida365-cli/internal/models"` — it already should since `ListProjects` uses it.

**Step 4: Remove `doRawRequest` from `internal/client/client.go`**

Delete the entire `doRawRequest` method (it was added in a previous commit and is no longer needed).

**Step 5: Update `runProjectData` in `cmd/project.go`**

Replace the current `runProjectData` function and clean up the now-unused imports:

```go
func runProjectData(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	c := loadClient()

	data, err := c.GetProjectData(projectID)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(data)
	return nil
}
```

Remove `"bytes"` and `"encoding/json"` from the imports in `cmd/project.go` — they are no longer used. The import block should become:

```go
import (
	"fmt"

	"github.com/bearzk/dida365-cli/internal/client"
	"github.com/bearzk/dida365-cli/internal/config"
	"github.com/spf13/cobra"
)
```

**Step 6: Run all tests**

```bash
go test ./...
```

Expected: All PASS. If there are compile errors about unused imports, fix them now.

**Step 7: Build and verify**

```bash
go build -o dida365 .
./dida365 project data --help
```

Expected: command still present, no panics

**Step 8: Commit**

```bash
git add internal/client/client.go internal/client/projects.go internal/client/projects_test.go cmd/project.go
git commit -m "refactor(client): replace raw GetProjectData with typed ProjectData response, remove doRawRequest"
```

---

## Task 3: Add `project columns` command

**Files:**
- Modify: `cmd/project.go`

**Context:** `GetProjectData` now returns `*models.ProjectData` which has a `Columns []Column` field. This command calls it and outputs just the columns slice — useful for discovering column IDs before moving tasks.

---

**Step 1: Add command definition and handler to `cmd/project.go`**

Add the command definition after `projectDataCmd`:

```go
var projectColumnsCmd = &cobra.Command{
	Use:   "columns <project-id>",
	Short: "List columns in a project",
	Long:  `List all Kanban columns for a project. Shows column IDs needed for 'task move'.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectColumns,
}
```

Add to `init()` after `projectCmd.AddCommand(projectDataCmd)`:

```go
projectCmd.AddCommand(projectColumnsCmd)
```

Add handler at end of file:

```go
func runProjectColumns(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	c := loadClient()

	data, err := c.GetProjectData(projectID)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(data.Columns)
	return nil
}
```

**Step 2: Build and verify**

```bash
go build -o dida365 .
./dida365 project --help
```

Expected output includes `columns` in Available Commands:
```
Available Commands:
  columns     List columns in a project
  data        Print raw project data response
  get         Get project details
  list        List all projects
```

**Step 3: Run all tests**

```bash
go test ./...
```

Expected: All PASS

**Step 4: Commit**

```bash
git add cmd/project.go
git commit -m "feat(cmd): add 'project columns' command to list Kanban columns"
```

---

## Task 4: Add `task move` command

**Files:**
- Modify: `cmd/task.go`

**Context:** Moving a task to a column is `UpdateTask(taskID, &TaskUpdate{ColumnID: &colID})`. The `TaskUpdate.ColumnID` field was added in Task 1. The `task move` command is an ergonomic wrapper — it sets only `ColumnID`, leaving all other fields untouched.

---

**Step 1: Add command definition, flag variable, and handler to `cmd/task.go`**

Add flag variable near the top with the other task flag variables:

```go
var taskColumnID string
```

Add command definition after `taskDeleteCmd`:

```go
var taskMoveCmd = &cobra.Command{
	Use:   "move <task-id>",
	Short: "Move a task to a different column",
	Long:  `Move a task to a different Kanban column. Use 'project columns <project-id>' to find column IDs.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskMove,
}
```

Add to `init()` after `taskCmd.AddCommand(taskDeleteCmd)`:

```go
taskCmd.AddCommand(taskMoveCmd)

// Flags for move command
taskMoveCmd.Flags().StringVar(&taskColumnID, "column-id", "", "Target column ID (required)")
taskMoveCmd.MarkFlagRequired("column-id")
```

Add handler at end of file:

```go
func runTaskMove(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	c := loadClient()

	updates := &models.TaskUpdate{
		ColumnID: &taskColumnID,
	}

	task, err := c.UpdateTask(taskID, updates)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	outputJSON(task)
	return nil
}
```

**Step 2: Build and verify**

```bash
go build -o dida365 .
./dida365 task --help
```

Expected output includes `move` in Available Commands:
```
Available Commands:
  complete    Mark a task as complete
  create      Create a new task
  delete      Delete a task
  get         Get task details
  list        List all tasks in a project
  move        Move a task to a different column
  update      Update a task
```

```bash
./dida365 task move --help
```

Expected: shows `--column-id` flag as required.

**Step 3: Run all tests**

```bash
go test ./...
```

Expected: All PASS

**Step 4: Commit**

```bash
git add cmd/task.go
git commit -m "feat(cmd): add 'task move' command to move tasks between Kanban columns"
```

---

## Summary

After all 4 tasks:

```bash
# See columns for a project (get the column IDs)
dida365 project columns <project-id>

# Move a task to a column
dida365 task move <task-id> --column-id <column-id>

# See full project data including tasks with columnId and columns array
dida365 project data <project-id>
```
