# Project Data Raw Command — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `dida365 project data <project-id>` command that prints the raw JSON response from `GET /open/v1/project/{projectId}/data` for API discovery.

**Architecture:** Add `doRawRequest` helper to the client (returns raw bytes, shares auth setup with `doRequest`), add `GetProjectData` method to `projects.go`, add `project data` subcommand to `cmd/project.go`.

**Tech Stack:** Go stdlib (encoding/json, net/http), Cobra CLI

---

## Task 1: Add `doRawRequest` and `GetProjectData` to the client

**Files:**
- Modify: `internal/client/client.go`
- Modify: `internal/client/projects.go`
- Modify: `internal/client/projects_test.go`

**Step 1: Write the failing test**

Add to `internal/client/projects_test.go`:

```go
func TestGetProjectData(t *testing.T) {
	t.Run("returns raw json bytes", func(t *testing.T) {
		rawResponse := `{"tasks":[{"id":"t1","title":"Test"}],"columns":[{"id":"c1","name":"To Do"}]}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("method: got %s, want GET", r.Method)
			}
			if r.URL.Path != "/open/v1/project/proj123/data" {
				t.Errorf("path: got %s, want /open/v1/project/proj123/data", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, rawResponse)
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		c := NewClient(cfg)
		result, err := c.GetProjectData("proj123")
		if err != nil {
			t.Fatalf("GetProjectData failed: %v", err)
		}

		if string(result) != rawResponse {
			t.Errorf("raw bytes: got %s, want %s", string(result), rawResponse)
		}
	})

	t.Run("returns error on api failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":"not found"}`)
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		c := NewClient(cfg)
		_, err := c.GetProjectData("missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
```

Note: Add `"fmt"` to the imports in `projects_test.go` if not already present.

**Step 2: Run the test to verify it fails**

```bash
go test ./internal/client -v -run TestGetProjectData
```

Expected: FAIL with `c.GetProjectData undefined`

**Step 3: Add `doRawRequest` to `internal/client/client.go`**

Add after the `doRequest` method (around line 88):

```go
// doRawRequest performs an HTTP GET request and returns the raw response body.
// Use this when you need the full JSON response without struct decoding.
func (c *Client) doRawRequest(path string) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.handleHTTPError(resp.StatusCode, body)
	}

	return body, nil
}
```

**Step 4: Add `GetProjectData` to `internal/client/projects.go`**

Add at the end of the file:

```go
// GetProjectData returns the raw JSON response from the project data endpoint.
// Use this to discover the full response shape including columns, tags, etc.
func (c *Client) GetProjectData(projectID string) ([]byte, error) {
	path := fmt.Sprintf("/open/v1/project/%s/data", projectID)
	data, err := c.doRawRequest(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get project data: %w", err)
	}
	return data, nil
}
```

Note: `encoding/json` is not needed here — the import block in `projects.go` already has `fmt`.

**Step 5: Run the test to verify it passes**

```bash
go test ./internal/client -v -run TestGetProjectData
```

Expected: PASS

**Step 6: Run the full test suite to check for regressions**

```bash
go test ./internal/client -v
```

Expected: All tests PASS

**Step 7: Commit**

```bash
git add internal/client/client.go internal/client/projects.go internal/client/projects_test.go
git commit -m "feat(client): add doRawRequest helper and GetProjectData method"
```

---

## Task 2: Add `project data` CLI command

**Files:**
- Modify: `cmd/project.go`

**Step 1: Add the command and handler to `cmd/project.go`**

Add the command definition after `projectGetCmd` (around line 30):

```go
var projectDataCmd = &cobra.Command{
	Use:   "data <project-id>",
	Short: "Print raw project data response",
	Long:  `Print the raw JSON response from GET /open/v1/project/{id}/data. Use this to inspect the full response shape including columns and other fields.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectData,
}
```

Add to `init()`, after `projectCmd.AddCommand(projectGetCmd)`:

```go
projectCmd.AddCommand(projectDataCmd)
```

Add the handler function at the end of the file:

```go
func runProjectData(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	c := loadClient()

	raw, err := c.GetProjectData(projectID)
	if err != nil {
		outputError(err, "API_ERROR", 3)
		return nil
	}

	// Pretty-print the raw JSON
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		// If indent fails (malformed JSON), print raw bytes anyway
		fmt.Println(string(raw))
		return nil
	}

	fmt.Println(buf.String())
	return nil
}
```

Add to imports at the top of `cmd/project.go`:

```go
import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bearzk/dida365-cli/internal/client"
	"github.com/bearzk/dida365-cli/internal/config"
	"github.com/spf13/cobra"
)
```

**Step 2: Build to verify it compiles**

```bash
go build -o dida365 .
```

Expected: No errors

**Step 3: Verify the command appears in help**

```bash
./dida365 project --help
```

Expected output includes:
```
Available Commands:
  data        Print raw project data response
  get         Get project details
  list        List all projects
```

**Step 4: Run all tests to check for regressions**

```bash
go test ./...
```

Expected: All tests PASS

**Step 5: Commit**

```bash
git add cmd/project.go
git commit -m "feat(cmd): add 'project data' command for raw API response inspection"
```

---

## Usage After Implementation

```bash
# See the full shape of any project's data endpoint
./dida365 project data <your-project-id>

# Example output (actual fields depend on API response):
{
  "tasks": [...],
  "columns": [
    { "id": "abc123", "name": "To Do", ... },
    { "id": "def456", "name": "In Progress", ... }
  ]
}
```

Once you see the real response, note:
- Whether `column_id` appears on task objects
- Whether a `columns` array is returned
- The exact field names for column data

This information feeds the next design phase: modeling columns and adding column-aware task operations.
