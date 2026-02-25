# Fix Time Parsing for Completed Tasks Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix CLI crash when reading completed tasks whose `completedTime` is in `+0000` (no-colon) timezone offset format.

**Architecture:** Add a `FlexTime` type in `internal/models/time.go` that implements `json.Unmarshaler`, falling back from RFC3339Nano to the ISO 8601 basic offset format. Replace `*time.Time` with `*FlexTime` in `models.Task`. All changes are isolated to the models package.

**Tech Stack:** Go standard library (`encoding/json`, `time`), existing `go test ./...` test suite.

---

### Task 1: Add FlexTime type with failing tests

**Files:**
- Create: `internal/models/time.go`
- Create: `internal/models/time_test.go`

**Step 1: Write the failing test**

Create `internal/models/time_test.go`:

```go
package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFlexTimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantUTC string // expected time in RFC3339 UTC
	}{
		{
			name:    "standard RFC3339Nano with Z",
			input:   `"2026-02-25T20:28:56.267Z"`,
			wantUTC: "2026-02-25T20:28:56.267Z",
		},
		{
			name:    "RFC3339 with colon offset",
			input:   `"2026-02-25T20:28:56.267+00:00"`,
			wantUTC: "2026-02-25T20:28:56.267Z",
		},
		{
			name:    "ISO 8601 basic offset without colon (Dida365 API format)",
			input:   `"2026-02-25T20:28:56.267+0000"`,
			wantUTC: "2026-02-25T20:28:56.267Z",
		},
		{
			name:    "ISO 8601 basic negative offset",
			input:   `"2026-02-25T15:28:56.267-0500"`,
			wantUTC: "2026-02-25T20:28:56.267Z",
		},
		{
			name:    "null JSON value",
			input:   `null`,
			wantErr: false,
			wantUTC: "",
		},
		{
			name:    "invalid format",
			input:   `"not-a-date"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ft FlexTime
			err := json.Unmarshal([]byte(tt.input), &ft)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantUTC == "" {
				// null input: zero value expected
				if !ft.IsZero() {
					t.Errorf("expected zero time, got %v", ft.Time)
				}
				return
			}

			got := ft.UTC().Format(time.RFC3339Nano)
			if got != tt.wantUTC {
				t.Errorf("got %s, want %s", got, tt.wantUTC)
			}
		})
	}
}

func TestTaskUnmarshalCompletedTime(t *testing.T) {
	// This is the exact payload the Dida365 API returns for a completed task
	input := `{
		"id": "abc123",
		"projectId": "proj456",
		"title": "Done task",
		"status": 2,
		"sortOrder": 0,
		"completedTime": "2026-02-25T20:28:56.267+0000"
	}`

	var task Task
	if err := json.Unmarshal([]byte(input), &task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.CompletedTime == nil {
		t.Fatal("expected completedTime to be set, got nil")
	}

	want := "2026-02-25T20:28:56.267Z"
	got := task.CompletedTime.UTC().Format(time.RFC3339Nano)
	if got != want {
		t.Errorf("completedTime: got %s, want %s", got, want)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/models/... -run TestFlexTime -v
go test ./internal/models/... -run TestTaskUnmarshalCompletedTime -v
```

Expected: FAIL — `FlexTime` undefined.

**Step 3: Write the FlexTime implementation**

Create `internal/models/time.go`:

```go
package models

import (
	"fmt"
	"time"
)

// FlexTime is a time.Time that can unmarshal both RFC3339 and ISO 8601 basic
// offset formats (e.g. "+0000" without colon), as returned by the Dida365 API
// for completedTime fields.
type FlexTime struct {
	time.Time
}

// flexTimeLayouts lists formats tried in order during JSON unmarshaling.
var flexTimeLayouts = []string{
	time.RFC3339Nano,                        // "2006-01-02T15:04:05.999999999Z07:00"
	"2006-01-02T15:04:05.999999999-0700",    // ISO 8601 basic: +0000 (no colon)
	time.RFC3339,                            // "2006-01-02T15:04:05Z07:00"
	"2006-01-02T15:04:05-0700",             // ISO 8601 basic without sub-seconds
}

// UnmarshalJSON implements json.Unmarshaler. It accepts null (zero value) and
// any of the formats in flexTimeLayouts.
func (ft *FlexTime) UnmarshalJSON(data []byte) error {
	s := string(data)

	if s == "null" {
		return nil
	}

	// Strip surrounding quotes
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return fmt.Errorf("FlexTime: expected JSON string, got %s", s)
	}
	s = s[1 : len(s)-1]

	for _, layout := range flexTimeLayouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			ft.Time = t
			return nil
		}
	}

	return fmt.Errorf("FlexTime: cannot parse %q as a time value", s)
}
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/models/... -run TestFlexTime -v
go test ./internal/models/... -run TestTaskUnmarshalCompletedTime -v
```

Expected: PASS for all subtests.

**Step 5: Commit**

```bash
git add internal/models/time.go internal/models/time_test.go
git commit -m "feat(models): add FlexTime type for ISO 8601 basic offset parsing"
```

---

### Task 2: Use FlexTime in Task model

**Files:**
- Modify: `internal/models/task.go` — change `CompletedTime *time.Time` to `*FlexTime`

**Step 1: Update the Task struct**

In `internal/models/task.go`, change:

```go
import "time"

// ...
CompletedTime *time.Time `json:"completedTime,omitempty"`
```

To:

```go
// (remove "time" import if no longer used elsewhere in file)

// ...
CompletedTime *FlexTime `json:"completedTime,omitempty"`
```

> Note: the `time` import can be removed entirely from `task.go` since `FlexTime` now lives in `time.go` (same package). Verify with `go build`.

**Step 2: Ensure existing tests still pass**

```bash
go test ./internal/models/... -v
```

Expected: All existing tests PASS. The `task_test.go` marshaling tests use `time.Now()` assigned to `*time.Time` — those will need updating since the field is now `*FlexTime`.

**Step 3: Fix task_test.go to use FlexTime**

In `internal/models/task_test.go`, replace all occurrences of `&now` (where `now` is `time.Time`) with a `FlexTime` value:

Change the declaration at the top of `TestTaskJSONMarshaling`:

```go
now := time.Now().UTC()
```

To:

```go
nowFlex := FlexTime{Time: time.Now().UTC()}
```

And replace all `&now` with `&nowFlex`.

**Step 4: Run full model tests**

```bash
go test ./internal/models/... -v
```

Expected: All tests PASS.

**Step 5: Run full test suite**

```bash
go test ./... -v
```

Expected: All tests PASS. No compilation errors.

**Step 6: Commit**

```bash
git add internal/models/task.go internal/models/task_test.go
git commit -m "fix(models): use FlexTime for Task.CompletedTime to handle +0000 offset"
```

---

### Task 3: Open PR referencing the issue

**Step 1: Push the branch**

```bash
git push -u origin HEAD
```

**Step 2: Create a PR**

```bash
gh pr create \
  --title "fix: handle +0000 timezone offset in completedTime (fixes #1)" \
  --body "$(cat <<'EOF'
## Summary

- Adds `FlexTime` type in `internal/models/time.go` that tries RFC3339Nano then ISO 8601 basic offset format when unmarshaling JSON
- Replaces `*time.Time` with `*FlexTime` on `Task.CompletedTime`
- Fixes crash reported in #1: `parsing time "2026-02-25T20:28:56.267+0000" as "2006-01-02T15:04:05Z07:00"`

## Test plan

- [ ] `TestFlexTimeUnmarshalJSON` covers RFC3339Nano, colon offset, no-colon offset, negative offset, null, and invalid input
- [ ] `TestTaskUnmarshalCompletedTime` uses the exact API payload from the bug report
- [ ] All existing model tests still pass
- [ ] Full `go test ./...` passes
EOF
)"
```

---
