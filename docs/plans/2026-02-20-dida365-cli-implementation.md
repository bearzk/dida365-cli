# Dida365 CLI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a production-ready Go CLI for Dida365 task management with full CRUD operations, focusing on scripting/automation use cases.

**Architecture:** Cobra CLI framework with clean layered architecture (cmd → internal/client → API), Viper for config management, JSON-only output for scripting. Test-driven development with httptest mock servers.

**Tech Stack:** Go 1.21+, Cobra, Viper, stdlib net/http, stdlib encoding/json

---

## Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `.gitignore`

**Step 1: Initialize Go module**

Run:
```bash
go mod init github.com/yourusername/dida365-cli
```

Expected: `go.mod` created with module declaration

**Step 2: Create main.go skeleton**

Create `main.go`:
```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("dida365 CLI")
	os.Exit(0)
}
```

**Step 3: Create .gitignore**

Create `.gitignore`:
```
# Binaries
dida365
dida365-cli
*.exe
*.dll
*.so
*.dylib

# Test binaries
*.test

# Output
*.out

# Go workspace
.DS_Store
vendor/

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# Test coverage
coverage.txt
coverage.html
```

**Step 4: Test build**

Run:
```bash
go build -o dida365 .
./dida365
```

Expected: Output "dida365 CLI" and exit 0

**Step 5: Commit**

```bash
git add go.mod main.go .gitignore
git commit -m "chore: initialize Go project structure"
```

---

## Task 2: Data Models - Project

**Files:**
- Create: `internal/models/project.go`
- Create: `internal/models/project_test.go`

**Step 1: Write JSON marshaling test for Project**

Create `internal/models/project_test.go`:
```go
package models

import (
	"encoding/json"
	"testing"
)

func TestProjectJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		project  Project
		expected string
	}{
		{
			name: "complete project",
			project: Project{
				ID:        "proj123",
				Name:      "Personal",
				Color:     "#FF0000",
				SortOrder: 1,
				Closed:    false,
				Kind:      "TASK",
			},
			expected: `{"id":"proj123","name":"Personal","color":"#FF0000","sortOrder":1,"closed":false,"kind":"TASK"}`,
		},
		{
			name: "project without optional color",
			project: Project{
				ID:        "proj456",
				Name:      "Work",
				SortOrder: 2,
				Closed:    true,
				Kind:      "NOTE",
			},
			expected: `{"id":"proj456","name":"Work","sortOrder":2,"closed":true,"kind":"NOTE"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.project)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}

			var expected map[string]interface{}
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			// Compare field by field
			for key, expectedVal := range expected {
				if result[key] != expectedVal {
					t.Errorf("field %s: got %v, want %v", key, result[key], expectedVal)
				}
			}
		})
	}
}

func TestProjectJSONUnmarshaling(t *testing.T) {
	jsonData := `{"id":"proj789","name":"Test Project","color":"#00FF00","sortOrder":3,"closed":false,"kind":"TASK"}`

	var project Project
	err := json.Unmarshal([]byte(jsonData), &project)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if project.ID != "proj789" {
		t.Errorf("ID: got %s, want proj789", project.ID)
	}
	if project.Name != "Test Project" {
		t.Errorf("Name: got %s, want Test Project", project.Name)
	}
	if project.Color != "#00FF00" {
		t.Errorf("Color: got %s, want #00FF00", project.Color)
	}
	if project.SortOrder != 3 {
		t.Errorf("SortOrder: got %d, want 3", project.SortOrder)
	}
	if project.Closed != false {
		t.Errorf("Closed: got %v, want false", project.Closed)
	}
	if project.Kind != "TASK" {
		t.Errorf("Kind: got %s, want TASK", project.Kind)
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/models -v
```

Expected: FAIL with "no such file or directory" or "undefined: Project"

**Step 3: Implement Project model**

Create `internal/models/project.go`:
```go
package models

// Project represents a Dida365 project/list
type Project struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color,omitempty"`
	SortOrder int    `json:"sortOrder"`
	Closed    bool   `json:"closed"`
	Kind      string `json:"kind"` // "TASK" or "NOTE"
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/models -v
```

Expected: PASS for all Project tests

**Step 5: Commit**

```bash
git add internal/models/
git commit -m "feat: add Project data model with JSON marshaling"
```

---

## Task 3: Data Models - Task

**Files:**
- Create: `internal/models/task.go`
- Create: `internal/models/task_test.go`

**Step 1: Write JSON marshaling test for Task**

Create `internal/models/task_test.go`:
```go
package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaskJSONMarshaling(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name string
		task Task
		checkFields map[string]interface{}
	}{
		{
			name: "complete task with all fields",
			task: Task{
				ID:            "task123",
				ProjectID:     "proj456",
				Title:         "Buy groceries",
				Content:       "Milk, eggs, bread",
				Status:        0,
				Priority:      3,
				CompletedTime: &now,
				SortOrder:     1,
			},
			checkFields: map[string]interface{}{
				"id":        "task123",
				"projectId": "proj456",
				"title":     "Buy groceries",
				"content":   "Milk, eggs, bread",
				"status":    float64(0),
				"priority":  float64(3),
				"sortOrder": float64(1),
			},
		},
		{
			name: "minimal task without optional fields",
			task: Task{
				ID:        "task789",
				ProjectID: "proj123",
				Title:     "Simple task",
				Status:    0,
				SortOrder: 0,
			},
			checkFields: map[string]interface{}{
				"id":        "task789",
				"projectId": "proj123",
				"title":     "Simple task",
				"status":    float64(0),
				"sortOrder": float64(0),
			},
		},
		{
			name: "completed task",
			task: Task{
				ID:            "task999",
				ProjectID:     "proj111",
				Title:         "Done task",
				Status:        2,
				CompletedTime: &now,
				SortOrder:     5,
			},
			checkFields: map[string]interface{}{
				"id":        "task999",
				"projectId": "proj111",
				"title":     "Done task",
				"status":    float64(2),
				"sortOrder": float64(5),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.task)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}

			for key, expected := range tt.checkFields {
				if result[key] != expected {
					t.Errorf("field %s: got %v, want %v", key, result[key], expected)
				}
			}

			// Check omitempty works
			if tt.task.Content == "" {
				if _, exists := result["content"]; exists {
					t.Error("empty content should be omitted")
				}
			}
		})
	}
}

func TestTaskCreateJSONMarshaling(t *testing.T) {
	tc := TaskCreate{
		Title:     "New task",
		ProjectID: "proj123",
		Content:   "Task description",
	}

	data, err := json.Marshal(tc)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["title"] != "New task" {
		t.Errorf("title: got %v, want New task", result["title"])
	}
	if result["projectId"] != "proj123" {
		t.Errorf("projectId: got %v, want proj123", result["projectId"])
	}
	if result["content"] != "Task description" {
		t.Errorf("content: got %v, want Task description", result["content"])
	}
}

func TestTaskUpdateJSONMarshaling(t *testing.T) {
	title := "Updated title"
	content := "Updated content"

	tu := TaskUpdate{
		Title:   &title,
		Content: &content,
	}

	data, err := json.Marshal(tu)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["title"] != "Updated title" {
		t.Errorf("title: got %v, want Updated title", result["title"])
	}
	if result["content"] != "Updated content" {
		t.Errorf("content: got %v, want Updated content", result["content"])
	}

	// Test omitempty with nil pointers
	tu2 := TaskUpdate{}
	data2, err := json.Marshal(tu2)
	if err != nil {
		t.Fatalf("failed to marshal empty update: %v", err)
	}

	var result2 map[string]interface{}
	if err := json.Unmarshal(data2, &result2); err != nil {
		t.Fatalf("failed to unmarshal empty: %v", err)
	}

	if len(result2) != 0 {
		t.Errorf("empty TaskUpdate should marshal to empty object, got %v", result2)
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/models -v
```

Expected: FAIL with "undefined: Task", "undefined: TaskCreate", etc.

**Step 3: Implement Task models**

Create `internal/models/task.go`:
```go
package models

import "time"

// Task represents a Dida365 task
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

// TaskCreate represents the payload for creating a new task
type TaskCreate struct {
	Title     string `json:"title"`
	ProjectID string `json:"projectId"`
	Content   string `json:"content,omitempty"`
}

// TaskUpdate represents the payload for updating a task
type TaskUpdate struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/models -v
```

Expected: PASS for all Task tests

**Step 5: Commit**

```bash
git add internal/models/task.go internal/models/task_test.go
git commit -m "feat: add Task data models with JSON marshaling"
```

---

## Task 4: Configuration Management

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write config load/save tests**

Create `internal/config/config_test.go`:
```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoad(t *testing.T) {
	t.Run("load valid config", func(t *testing.T) {
		// Create temp directory
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		// Write valid config
		validConfig := Config{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			BaseURL:      "https://dida365.com",
		}
		data, _ := json.MarshalIndent(validConfig, "", "  ")
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		// Load config
		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.ClientID != "test_client_id" {
			t.Errorf("ClientID: got %s, want test_client_id", cfg.ClientID)
		}
		if cfg.ClientSecret != "test_client_secret" {
			t.Errorf("ClientSecret: got %s, want test_client_secret", cfg.ClientSecret)
		}
		if cfg.AccessToken != "test_access_token" {
			t.Errorf("AccessToken: got %s, want test_access_token", cfg.AccessToken)
		}
		if cfg.BaseURL != "https://dida365.com" {
			t.Errorf("BaseURL: got %s, want https://dida365.com", cfg.BaseURL)
		}
	})

	t.Run("load nonexistent config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nonexistent.json")

		_, err := Load(configPath)
		if err == nil {
			t.Fatal("expected error for nonexistent config, got nil")
		}
	})

	t.Run("load invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		// Write invalid JSON
		if err := os.WriteFile(configPath, []byte("{invalid json}"), 0600); err != nil {
			t.Fatalf("failed to write invalid config: %v", err)
		}

		_, err := Load(configPath)
		if err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})
}

func TestConfigSave(t *testing.T) {
	t.Run("save valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		cfg := &Config{
			ClientID:     "save_test_id",
			ClientSecret: "save_test_secret",
			AccessToken:  "save_test_token",
			BaseURL:      "https://dida365.com",
		}

		if err := cfg.Save(configPath); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify file exists and has correct permissions
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("config file not created: %v", err)
		}

		if info.Mode().Perm() != 0600 {
			t.Errorf("config file permissions: got %o, want 0600", info.Mode().Perm())
		}

		// Verify content
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		var loaded Config
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("failed to unmarshal saved config: %v", err)
		}

		if loaded.ClientID != "save_test_id" {
			t.Errorf("ClientID: got %s, want save_test_id", loaded.ClientID)
		}
	})

	t.Run("save creates directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedPath := filepath.Join(tmpDir, "nested", "dir", "config.json")

		cfg := &Config{
			ClientID:     "nested_test",
			ClientSecret: "nested_secret",
			AccessToken:  "nested_token",
			BaseURL:      "https://dida365.com",
		}

		if err := cfg.Save(nestedPath); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(nestedPath); err != nil {
			t.Errorf("config file not created in nested path: %v", err)
		}
	})
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "valid config",
			config: Config{
				ClientID:     "valid_id",
				ClientSecret: "valid_secret",
				AccessToken:  "valid_token",
				BaseURL:      "https://dida365.com",
			},
			wantError: false,
		},
		{
			name: "missing client_id",
			config: Config{
				ClientSecret: "valid_secret",
				AccessToken:  "valid_token",
				BaseURL:      "https://dida365.com",
			},
			wantError: true,
		},
		{
			name: "missing client_secret",
			config: Config{
				ClientID:    "valid_id",
				AccessToken: "valid_token",
				BaseURL:     "https://dida365.com",
			},
			wantError: true,
		},
		{
			name: "missing access_token",
			config: Config{
				ClientID:     "valid_id",
				ClientSecret: "valid_secret",
				BaseURL:      "https://dida365.com",
			},
			wantError: true,
		},
		{
			name: "missing base_url",
			config: Config{
				ClientID:     "valid_id",
				ClientSecret: "valid_secret",
				AccessToken:  "valid_token",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Error("DefaultConfigPath returned empty string")
	}

	// Should contain .dida365/config.json
	if !filepath.IsAbs(path) {
		t.Errorf("DefaultConfigPath should return absolute path, got %s", path)
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/config -v
```

Expected: FAIL with "undefined: Config", "undefined: Load", etc.

**Step 3: Implement Config package**

Create `internal/config/config.go`:
```go
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the CLI configuration
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AccessToken  string `json:"access_token"`
	BaseURL      string `json:"base_url"`
}

// Load reads config from the specified path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &cfg, nil
}

// Save writes config to the specified path
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with user-only permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if all required fields are present
func (c *Config) Validate() error {
	if c.ClientID == "" {
		return errors.New("client_id is required")
	}
	if c.ClientSecret == "" {
		return errors.New("client_secret is required")
	}
	if c.AccessToken == "" {
		return errors.New("access_token is required")
	}
	if c.BaseURL == "" {
		return errors.New("base_url is required")
	}
	return nil
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".dida365", "config.json")
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/config -v
```

Expected: PASS for all config tests

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config management with load/save/validate"
```

---

## Task 5: HTTP Client Foundation

**Files:**
- Create: `internal/client/client.go`
- Create: `internal/client/client_test.go`

**Step 1: Write HTTP client tests**

Create `internal/client/client_test.go`:
```go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/dida365-cli/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		ClientID:     "test_id",
		ClientSecret: "test_secret",
		AccessToken:  "test_token",
		BaseURL:      "https://dida365.com",
	}

	client := NewClient(cfg)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.config != cfg {
		t.Error("client config not set correctly")
	}

	if client.baseURL != "https://dida365.com" {
		t.Errorf("baseURL: got %s, want https://dida365.com", client.baseURL)
	}
}

func TestDoRequestAuthHeader(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test_token_123" {
			t.Errorf("Authorization header: got %s, want Bearer test_token_123", auth)
		}

		// Check content-type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type header: got %s, want application/json", contentType)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	cfg := &config.Config{
		ClientID:     "test_id",
		ClientSecret: "test_secret",
		AccessToken:  "test_token_123",
		BaseURL:      server.URL,
	}

	client := NewClient(cfg)

	var result map[string]string
	err := client.doRequest("GET", "/test", nil, &result)
	if err != nil {
		t.Fatalf("doRequest failed: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("response status: got %s, want ok", result["status"])
	}
}

func TestDoRequestHTTPErrors(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   map[string]interface{}
		expectedErrMsg string
	}{
		{
			name:           "401 unauthorized",
			statusCode:     http.StatusUnauthorized,
			responseBody:   map[string]interface{}{"error": "invalid token"},
			expectedErrMsg: "access token expired or invalid",
		},
		{
			name:           "403 forbidden",
			statusCode:     http.StatusForbidden,
			responseBody:   map[string]interface{}{"error": "insufficient permissions"},
			expectedErrMsg: "insufficient permissions for this operation",
		},
		{
			name:           "404 not found",
			statusCode:     http.StatusNotFound,
			responseBody:   map[string]interface{}{"error": "not found"},
			expectedErrMsg: "resource not found",
		},
		{
			name:           "500 server error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   map[string]interface{}{"error": "internal error"},
			expectedErrMsg: "Dida365 server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			cfg := &config.Config{
				ClientID:     "test_id",
				ClientSecret: "test_secret",
				AccessToken:  "test_token",
				BaseURL:      server.URL,
			}

			client := NewClient(cfg)

			var result map[string]string
			err := client.doRequest("GET", "/test", nil, &result)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("error message: got %q, want to contain %q", err.Error(), tt.expectedErrMsg)
			}
		})
	}
}

func TestDoRequestJSONMarshaling(t *testing.T) {
	t.Run("request body marshaling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			if body["field"] != "value" {
				t.Errorf("request body field: got %s, want value", body["field"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"result": "success"})
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test_id",
			ClientSecret: "test_secret",
			AccessToken:  "test_token",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)

		requestBody := map[string]string{"field": "value"}
		var result map[string]string
		err := client.doRequest("POST", "/test", requestBody, &result)
		if err != nil {
			t.Fatalf("doRequest failed: %v", err)
		}

		if result["result"] != "success" {
			t.Errorf("response result: got %s, want success", result["result"])
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/client -v
```

Expected: FAIL with "undefined: Client", "undefined: NewClient", etc.

**Step 3: Implement HTTP client**

Create `internal/client/client.go`:
```go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/yourusername/dida365-cli/internal/config"
)

// Client is the HTTP client for Dida365 API
type Client struct {
	httpClient *http.Client
	config     *config.Config
	baseURL    string
}

// NewClient creates a new API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{},
		config:     cfg,
		baseURL:    cfg.BaseURL,
	}
}

// doRequest performs an HTTP request with auth headers
func (c *Client) doRequest(method, path string, body, result interface{}) error {
	// Build full URL
	url := c.baseURL + path

	// Marshal request body if provided
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	}

	// Create request
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return c.handleHTTPError(resp.StatusCode, respBody)
	}

	// Unmarshal successful response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// handleHTTPError converts HTTP error responses to Go errors
func (c *Client) handleHTTPError(statusCode int, body []byte) error {
	var apiError map[string]interface{}
	_ = json.Unmarshal(body, &apiError)

	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("access token expired or invalid")
	case http.StatusForbidden:
		return fmt.Errorf("insufficient permissions for this operation")
	case http.StatusNotFound:
		return fmt.Errorf("resource not found")
	case http.StatusBadRequest:
		if msg, ok := apiError["error"].(string); ok {
			return fmt.Errorf("bad request: %s", msg)
		}
		return fmt.Errorf("bad request")
	case http.StatusInternalServerError:
		return fmt.Errorf("Dida365 server error, try again later")
	default:
		return fmt.Errorf("HTTP %d: %s", statusCode, string(body))
	}
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/client -v
```

Expected: PASS for all client tests

**Step 5: Update go.mod**

Run:
```bash
go mod tidy
```

**Step 6: Commit**

```bash
git add internal/client/ go.mod go.sum
git commit -m "feat: add HTTP client with auth and error handling"
```

---

## Task 6: Project API Operations

**Files:**
- Create: `internal/client/projects.go`
- Create: `internal/client/projects_test.go`

**Step 1: Write project API tests**

Create `internal/client/projects_test.go`:
```go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/dida365-cli/internal/config"
	"github.com/yourusername/dida365-cli/internal/models"
)

func TestListProjects(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		projects := []models.Project{
			{
				ID:        "proj1",
				Name:      "Personal",
				Color:     "#FF0000",
				SortOrder: 1,
				Closed:    false,
				Kind:      "TASK",
			},
			{
				ID:        "proj2",
				Name:      "Work",
				Color:     "#00FF00",
				SortOrder: 2,
				Closed:    false,
				Kind:      "TASK",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("method: got %s, want GET", r.Method)
			}
			if r.URL.Path != "/open/v1/project" {
				t.Errorf("path: got %s, want /open/v1/project", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(projects)
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)
		result, err := client.ListProjects()
		if err != nil {
			t.Fatalf("ListProjects failed: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected 2 projects, got %d", len(result))
		}

		if result[0].ID != "proj1" {
			t.Errorf("project 0 ID: got %s, want proj1", result[0].ID)
		}
		if result[1].Name != "Work" {
			t.Errorf("project 1 Name: got %s, want Work", result[1].Name)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]models.Project{})
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)
		result, err := client.ListProjects()
		if err != nil {
			t.Fatalf("ListProjects failed: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("expected empty list, got %d projects", len(result))
		}
	})

	t.Run("unauthorized error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "invalid",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)
		_, err := client.ListProjects()
		if err == nil {
			t.Fatal("expected error for unauthorized, got nil")
		}

		if !contains(err.Error(), "access token") {
			t.Errorf("error message should mention access token, got: %s", err.Error())
		}
	})
}

func TestGetProject(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		project := models.Project{
			ID:        "proj123",
			Name:      "Test Project",
			Color:     "#0000FF",
			SortOrder: 5,
			Closed:    false,
			Kind:      "TASK",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("method: got %s, want GET", r.Method)
			}
			if r.URL.Path != "/open/v1/project/proj123" {
				t.Errorf("path: got %s, want /open/v1/project/proj123", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(project)
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)
		result, err := client.GetProject("proj123")
		if err != nil {
			t.Fatalf("GetProject failed: %v", err)
		}

		if result.ID != "proj123" {
			t.Errorf("ID: got %s, want proj123", result.ID)
		}
		if result.Name != "Test Project" {
			t.Errorf("Name: got %s, want Test Project", result.Name)
		}
	})

	t.Run("not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}))
		defer server.Close()

		cfg := &config.Config{
			ClientID:     "test",
			ClientSecret: "test",
			AccessToken:  "test",
			BaseURL:      server.URL,
		}

		client := NewClient(cfg)
		_, err := client.GetProject("nonexistent")
		if err == nil {
			t.Fatal("expected error for not found, got nil")
		}

		if !contains(err.Error(), "not found") {
			t.Errorf("error should mention not found, got: %s", err.Error())
		}
	})
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/client -v
```

Expected: FAIL with "undefined: ListProjects", "undefined: GetProject"

**Step 3: Implement project operations**

Create `internal/client/projects.go`:
```go
package client

import (
	"fmt"

	"github.com/yourusername/dida365-cli/internal/models"
)

// ListProjects retrieves all projects for the authenticated user
func (c *Client) ListProjects() ([]*models.Project, error) {
	var projects []*models.Project
	if err := c.doRequest("GET", "/open/v1/project", nil, &projects); err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}

// GetProject retrieves a specific project by ID
func (c *Client) GetProject(projectID string) (*models.Project, error) {
	var project models.Project
	path := fmt.Sprintf("/open/v1/project/%s", projectID)
	if err := c.doRequest("GET", path, nil, &project); err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/client -v
```

Expected: PASS for all tests including project operations

**Step 5: Commit**

```bash
git add internal/client/projects.go internal/client/projects_test.go
git commit -m "feat: add project API operations (list, get)"
```

---

## Task 7: Task API Operations

**Files:**
- Create: `internal/client/tasks.go`
- Create: `internal/client/tasks_test.go`

**Step 1: Write task API tests**

Create `internal/client/tasks_test.go`:
```go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/dida365-cli/internal/config"
	"github.com/yourusername/dida365-cli/internal/models"
)

func TestCreateTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method: got %s, want POST", r.Method)
		}
		if r.URL.Path != "/open/v1/task" {
			t.Errorf("path: got %s, want /open/v1/task", r.URL.Path)
		}

		var req models.TaskCreate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Title != "Test Task" {
			t.Errorf("title: got %s, want Test Task", req.Title)
		}
		if req.ProjectID != "proj123" {
			t.Errorf("projectId: got %s, want proj123", req.ProjectID)
		}

		task := models.Task{
			ID:        "task456",
			ProjectID: req.ProjectID,
			Title:     req.Title,
			Content:   req.Content,
			Status:    0,
			SortOrder: 0,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}))
	defer server.Close()

	cfg := &config.Config{
		ClientID:     "test",
		ClientSecret: "test",
		AccessToken:  "test",
		BaseURL:      server.URL,
	}

	client := NewClient(cfg)
	taskCreate := &models.TaskCreate{
		Title:     "Test Task",
		ProjectID: "proj123",
		Content:   "Task description",
	}

	result, err := client.CreateTask(taskCreate)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if result.ID != "task456" {
		t.Errorf("ID: got %s, want task456", result.ID)
	}
	if result.Title != "Test Task" {
		t.Errorf("Title: got %s, want Test Task", result.Title)
	}
}

func TestGetTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method: got %s, want GET", r.Method)
		}
		if r.URL.Path != "/open/v1/project/proj123/task/task456" {
			t.Errorf("path: got %s, want /open/v1/project/proj123/task/task456", r.URL.Path)
		}

		task := models.Task{
			ID:        "task456",
			ProjectID: "proj123",
			Title:     "Existing Task",
			Status:    0,
			SortOrder: 1,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(task)
	}))
	defer server.Close()

	cfg := &config.Config{
		ClientID:     "test",
		ClientSecret: "test",
		AccessToken:  "test",
		BaseURL:      server.URL,
	}

	client := NewClient(cfg)
	result, err := client.GetTask("proj123", "task456")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if result.ID != "task456" {
		t.Errorf("ID: got %s, want task456", result.ID)
	}
	if result.Title != "Existing Task" {
		t.Errorf("Title: got %s, want Existing Task", result.Title)
	}
}

func TestListTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method: got %s, want GET", r.Method)
		}
		if r.URL.Path != "/open/v1/project/proj123/data" {
			t.Errorf("path: got %s, want /open/v1/project/proj123/data", r.URL.Path)
		}

		// API returns project data with tasks array
		response := map[string]interface{}{
			"id":   "proj123",
			"name": "Test Project",
			"tasks": []models.Task{
				{ID: "task1", ProjectID: "proj123", Title: "Task 1", Status: 0, SortOrder: 1},
				{ID: "task2", ProjectID: "proj123", Title: "Task 2", Status: 2, SortOrder: 2},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := &config.Config{
		ClientID:     "test",
		ClientSecret: "test",
		AccessToken:  "test",
		BaseURL:      server.URL,
	}

	client := NewClient(cfg)
	result, err := client.ListTasks("proj123")
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result))
	}

	if result[0].ID != "task1" {
		t.Errorf("task 0 ID: got %s, want task1", result[0].ID)
	}
	if result[1].Title != "Task 2" {
		t.Errorf("task 1 Title: got %s, want Task 2", result[1].Title)
	}
}

func TestUpdateTask(t *testing.T) {
	newTitle := "Updated Title"
	newContent := "Updated Content"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method: got %s, want POST", r.Method)
		}
		if r.URL.Path != "/open/v1/task/task456" {
			t.Errorf("path: got %s, want /open/v1/task/task456", r.URL.Path)
		}

		var req models.TaskUpdate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Title == nil || *req.Title != "Updated Title" {
			t.Error("title not updated correctly")
		}

		task := models.Task{
			ID:        "task456",
			ProjectID: "proj123",
			Title:     newTitle,
			Content:   newContent,
			Status:    0,
			SortOrder: 1,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(task)
	}))
	defer server.Close()

	cfg := &config.Config{
		ClientID:     "test",
		ClientSecret: "test",
		AccessToken:  "test",
		BaseURL:      server.URL,
	}

	client := NewClient(cfg)
	updates := &models.TaskUpdate{
		Title:   &newTitle,
		Content: &newContent,
	}

	result, err := client.UpdateTask("proj123", "task456", updates)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	if result.Title != "Updated Title" {
		t.Errorf("Title: got %s, want Updated Title", result.Title)
	}
}

func TestCompleteTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method: got %s, want POST", r.Method)
		}
		if r.URL.Path != "/open/v1/project/proj123/task/task456/complete" {
			t.Errorf("path: got %s, want /open/v1/project/proj123/task/task456/complete", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	cfg := &config.Config{
		ClientID:     "test",
		ClientSecret: "test",
		AccessToken:  "test",
		BaseURL:      server.URL,
	}

	client := NewClient(cfg)
	err := client.CompleteTask("proj123", "task456")
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}
}

func TestDeleteTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method: got %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/open/v1/project/proj123/task/task456" {
			t.Errorf("path: got %s, want /open/v1/project/proj123/task/task456", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	cfg := &config.Config{
		ClientID:     "test",
		ClientSecret: "test",
		AccessToken:  "test",
		BaseURL:      server.URL,
	}

	client := NewClient(cfg)
	err := client.DeleteTask("proj123", "task456")
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/client -v -run TestCreateTask
```

Expected: FAIL with "undefined: CreateTask"

**Step 3: Implement task operations**

Create `internal/client/tasks.go`:
```go
package client

import (
	"fmt"

	"github.com/yourusername/dida365-cli/internal/models"
)

// CreateTask creates a new task
func (c *Client) CreateTask(task *models.TaskCreate) (*models.Task, error) {
	var result models.Task
	if err := c.doRequest("POST", "/open/v1/task", task, &result); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return &result, nil
}

// GetTask retrieves a specific task
func (c *Client) GetTask(projectID, taskID string) (*models.Task, error) {
	var task models.Task
	path := fmt.Sprintf("/open/v1/project/%s/task/%s", projectID, taskID)
	if err := c.doRequest("GET", path, nil, &task); err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// ListTasks retrieves all tasks in a project
func (c *Client) ListTasks(projectID string) ([]*models.Task, error) {
	path := fmt.Sprintf("/open/v1/project/%s/data", projectID)

	// API returns project data with tasks embedded
	var response struct {
		Tasks []*models.Task `json:"tasks"`
	}

	if err := c.doRequest("GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return response.Tasks, nil
}

// UpdateTask updates an existing task
func (c *Client) UpdateTask(projectID, taskID string, updates *models.TaskUpdate) (*models.Task, error) {
	var result models.Task
	path := fmt.Sprintf("/open/v1/task/%s", taskID)
	if err := c.doRequest("POST", path, updates, &result); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	return &result, nil
}

// CompleteTask marks a task as complete
func (c *Client) CompleteTask(projectID, taskID string) error {
	path := fmt.Sprintf("/open/v1/project/%s/task/%s/complete", projectID, taskID)
	if err := c.doRequest("POST", path, nil, nil); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}
	return nil
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(projectID, taskID string) error {
	path := fmt.Sprintf("/open/v1/project/%s/task/%s", projectID, taskID)
	if err := c.doRequest("DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/client -v
```

Expected: PASS for all client tests

**Step 5: Commit**

```bash
git add internal/client/tasks.go internal/client/tasks_test.go
git commit -m "feat: add task API operations (CRUD + complete)"
```

---

## Task 8: Cobra CLI Setup

**Files:**
- Modify: `main.go`
- Create: `cmd/root.go`

**Step 1: Install Cobra**

Run:
```bash
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
```

**Step 2: Create root command**

Create `cmd/root.go`:
```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dida365",
	Short: "CLI for Dida365 task management",
	Long:  `A command-line interface for Dida365 (滴答清单) task management, designed for automation and scripting workflows.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = "0.1.0"
}
```

**Step 3: Update main.go**

Modify `main.go`:
```go
package main

import "github.com/yourusername/dida365-cli/cmd"

func main() {
	cmd.Execute()
}
```

**Step 4: Test CLI**

Run:
```bash
go build -o dida365 .
./dida365 --help
./dida365 --version
```

Expected: Help text displays with "dida365" usage, version shows "0.1.0"

**Step 5: Commit**

```bash
go mod tidy
git add main.go cmd/root.go go.mod go.sum
git commit -m "feat: add Cobra CLI framework with root command"
```

---

## Task 9: Auth Commands

**Files:**
- Create: `cmd/auth.go`

**Step 1: Implement auth commands**

Create `cmd/auth.go`:
```go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/dida365-cli/internal/client"
	"github.com/yourusername/dida365-cli/internal/config"
)

var (
	clientID     string
	clientSecret string
	accessToken  string
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication configuration",
	Long:  `Configure and validate Dida365 API credentials.`,
}

var authConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure API credentials",
	Long:  `Set client ID, client secret, and access token for API authentication.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create config
		cfg := &config.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			AccessToken:  accessToken,
			BaseURL:      "https://dida365.com",
		}

		// Validate required fields
		if err := cfg.Validate(); err != nil {
			return outputError(err, "VALIDATION_ERROR", 5)
		}

		// Test connection
		apiClient := client.NewClient(cfg)
		if _, err := apiClient.ListProjects(); err != nil {
			return outputError(fmt.Errorf("failed to validate token: %w", err), "CONNECTION_ERROR", 2)
		}

		// Save config
		configPath := config.DefaultConfigPath()
		if err := cfg.Save(configPath); err != nil {
			return outputError(err, "SAVE_ERROR", 1)
		}

		// Output success
		result := map[string]interface{}{
			"configured":  true,
			"config_path": configPath,
		}
		return outputJSON(result)
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Verify configuration exists and token is valid.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.DefaultConfigPath()
		cfg, err := config.Load(configPath)
		if err != nil {
			result := map[string]interface{}{
				"configured":  false,
				"token_valid": false,
				"error":       "configuration not found",
			}
			outputJSON(result)
			os.Exit(1)
			return nil
		}

		// Validate config
		if err := cfg.Validate(); err != nil {
			result := map[string]interface{}{
				"configured":  false,
				"token_valid": false,
				"error":       err.Error(),
			}
			outputJSON(result)
			os.Exit(1)
			return nil
		}

		// Test token
		apiClient := client.NewClient(cfg)
		_, err = apiClient.ListProjects()
		tokenValid := err == nil

		result := map[string]interface{}{
			"configured":  true,
			"client_id":   cfg.ClientID,
			"token_valid": tokenValid,
		}

		if !tokenValid {
			result["error"] = err.Error()
			outputJSON(result)
			os.Exit(2)
			return nil
		}

		return outputJSON(result)
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authConfigureCmd)
	authCmd.AddCommand(authStatusCmd)

	authConfigureCmd.Flags().StringVar(&clientID, "client-id", "", "Client ID from Dida365 developer portal (required)")
	authConfigureCmd.Flags().StringVar(&clientSecret, "client-secret", "", "Client secret from Dida365 developer portal (required)")
	authConfigureCmd.Flags().StringVar(&accessToken, "access-token", "", "Access token from OAuth flow (required)")
	authConfigureCmd.MarkFlagRequired("client-id")
	authConfigureCmd.MarkFlagRequired("client-secret")
	authConfigureCmd.MarkFlagRequired("access-token")
}

// Helper functions for output
func outputJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func outputError(err error, code string, exitCode int) error {
	errObj := map[string]interface{}{
		"error": err.Error(),
		"code":  code,
	}
	encoder := json.NewEncoder(os.Stderr)
	encoder.SetIndent("", "  ")
	encoder.Encode(errObj)
	os.Exit(exitCode)
	return nil
}
```

**Step 2: Test auth commands**

Run:
```bash
go build -o dida365 .
./dida365 auth --help
./dida365 auth configure --help
./dida365 auth status
```

Expected:
- Help text displays correctly
- `auth status` returns JSON error (config not found)

**Step 3: Commit**

```bash
git add cmd/auth.go
git commit -m "feat: add auth configure and auth status commands"
```

---

## Task 10: Project Commands

**Files:**
- Create: `cmd/project.go`

**Step 1: Implement project commands**

Create `cmd/project.go`:
```go
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yourusername/dida365-cli/internal/client"
	"github.com/yourusername/dida365-cli/internal/config"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  `List and view Dida365 projects.`,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  `Retrieve and display all projects for the authenticated user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := loadClient()
		if err != nil {
			return err
		}

		projects, err := apiClient.ListProjects()
		if err != nil {
			return outputError(err, "API_ERROR", 3)
		}

		return outputJSON(projects)
	},
}

var projectGetCmd = &cobra.Command{
	Use:   "get <project-id>",
	Short: "Get project details",
	Long:  `Retrieve detailed information about a specific project.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		apiClient, err := loadClient()
		if err != nil {
			return err
		}

		project, err := apiClient.GetProject(projectID)
		if err != nil {
			return outputError(err, "API_ERROR", 3)
		}

		return outputJSON(project)
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectGetCmd)
}

// loadClient loads config and creates API client
func loadClient() (*client.Client, error) {
	configPath := config.DefaultConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, outputError(err, "CONFIG_ERROR", 1)
	}

	if err := cfg.Validate(); err != nil {
		return nil, outputError(err, "CONFIG_ERROR", 1)
	}

	return client.NewClient(cfg), nil
}
```

**Step 2: Test project commands**

Run:
```bash
go build -o dida365 .
./dida365 project --help
./dida365 project list
./dida365 project get --help
```

Expected:
- Help text displays
- `project list` returns config error (not configured)

**Step 3: Commit**

```bash
git add cmd/project.go
git commit -m "feat: add project list and project get commands"
```

---

## Task 11: Task Commands

**Files:**
- Create: `cmd/task.go`

**Step 1: Implement task commands**

Create `cmd/task.go`:
```go
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yourusername/dida365-cli/internal/models"
)

var (
	taskTitle     string
	taskProjectID string
	taskContent   string
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  `Create, read, update, delete, and complete tasks.`,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long:  `Create a new task in the specified project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := loadClient()
		if err != nil {
			return err
		}

		taskCreate := &models.TaskCreate{
			Title:     taskTitle,
			ProjectID: taskProjectID,
			Content:   taskContent,
		}

		task, err := apiClient.CreateTask(taskCreate)
		if err != nil {
			return outputError(err, "API_ERROR", 3)
		}

		return outputJSON(task)
	},
}

var taskGetCmd = &cobra.Command{
	Use:   "get <task-id>",
	Short: "Get task details",
	Long:  `Retrieve detailed information about a specific task.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		apiClient, err := loadClient()
		if err != nil {
			return err
		}

		task, err := apiClient.GetTask(taskProjectID, taskID)
		if err != nil {
			return outputError(err, "API_ERROR", 3)
		}

		return outputJSON(task)
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list <project-id>",
	Short: "List tasks in a project",
	Long:  `Retrieve all tasks in the specified project.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		apiClient, err := loadClient()
		if err != nil {
			return err
		}

		tasks, err := apiClient.ListTasks(projectID)
		if err != nil {
			return outputError(err, "API_ERROR", 3)
		}

		return outputJSON(tasks)
	},
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update <task-id>",
	Short: "Update a task",
	Long:  `Update title and/or content of an existing task.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		apiClient, err := loadClient()
		if err != nil {
			return err
		}

		updates := &models.TaskUpdate{}

		// Only include fields that were provided
		if cmd.Flags().Changed("title") {
			updates.Title = &taskTitle
		}
		if cmd.Flags().Changed("content") {
			updates.Content = &taskContent
		}

		task, err := apiClient.UpdateTask(taskProjectID, taskID, updates)
		if err != nil {
			return outputError(err, "API_ERROR", 3)
		}

		return outputJSON(task)
	},
}

var taskCompleteCmd = &cobra.Command{
	Use:   "complete <task-id>",
	Short: "Mark task as complete",
	Long:  `Mark the specified task as completed.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		apiClient, err := loadClient()
		if err != nil {
			return err
		}

		if err := apiClient.CompleteTask(taskProjectID, taskID); err != nil {
			return outputError(err, "API_ERROR", 3)
		}

		result := map[string]interface{}{
			"status":  "completed",
			"task_id": taskID,
		}
		return outputJSON(result)
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete <task-id>",
	Short: "Delete a task",
	Long:  `Permanently delete the specified task.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		apiClient, err := loadClient()
		if err != nil {
			return err
		}

		if err := apiClient.DeleteTask(taskProjectID, taskID); err != nil {
			return outputError(err, "API_ERROR", 3)
		}

		result := map[string]interface{}{
			"status":  "deleted",
			"task_id": taskID,
		}
		return outputJSON(result)
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskGetCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	taskCmd.AddCommand(taskDeleteCmd)

	// Create flags
	taskCreateCmd.Flags().StringVar(&taskTitle, "title", "", "Task title (required)")
	taskCreateCmd.Flags().StringVar(&taskProjectID, "project-id", "", "Project ID (required)")
	taskCreateCmd.Flags().StringVar(&taskContent, "content", "", "Task content/description")
	taskCreateCmd.MarkFlagRequired("title")
	taskCreateCmd.MarkFlagRequired("project-id")

	// Get flags
	taskGetCmd.Flags().StringVar(&taskProjectID, "project-id", "", "Project ID (required)")
	taskGetCmd.MarkFlagRequired("project-id")

	// Update flags
	taskUpdateCmd.Flags().StringVar(&taskTitle, "title", "", "New task title")
	taskUpdateCmd.Flags().StringVar(&taskContent, "content", "", "New task content")
	taskUpdateCmd.Flags().StringVar(&taskProjectID, "project-id", "", "Project ID (required)")
	taskUpdateCmd.MarkFlagRequired("project-id")

	// Complete flags
	taskCompleteCmd.Flags().StringVar(&taskProjectID, "project-id", "", "Project ID (required)")
	taskCompleteCmd.MarkFlagRequired("project-id")

	// Delete flags
	taskDeleteCmd.Flags().StringVar(&taskProjectID, "project-id", "", "Project ID (required)")
	taskDeleteCmd.MarkFlagRequired("project-id")
}
```

**Step 2: Test task commands**

Run:
```bash
go build -o dida365 .
./dida365 task --help
./dida365 task create --help
./dida365 task get --help
./dida365 task list --help
```

Expected: All help text displays correctly with required flags

**Step 3: Commit**

```bash
git add cmd/task.go
git commit -m "feat: add task commands (create, get, list, update, complete, delete)"
```

---

## Task 12: README Documentation

**Files:**
- Create: `README.md`

**Step 1: Write README**

Create `README.md`:
```markdown
# Dida365 CLI

A command-line interface for [Dida365](https://dida365.com) (滴答清单 / TickTick) task management, designed for automation and scripting workflows.

## Features

- **Task Management**: Full CRUD operations for tasks (create, read, update, delete, complete)
- **Project Access**: List and view projects
- **JSON Output**: All commands output JSON for easy parsing with `jq` or other tools
- **Scripting-Friendly**: Clear exit codes and structured error messages
- **Secure**: Config file stored with user-only permissions

## Installation

### From Source

```bash
git clone https://github.com/yourusername/dida365-cli.git
cd dida365-cli
go build -o dida365 .
sudo mv dida365 /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/yourusername/dida365-cli@latest
```

## Getting Started

### 1. Obtain API Credentials

1. Visit the [Dida365 Developer Portal](https://developer.dida365.com)
2. Register an application to get `client_id` and `client_secret`
3. Complete the OAuth flow to obtain an `access_token`

### 2. Configure CLI

```bash
dida365 auth configure \
  --client-id "your_client_id" \
  --client-secret "your_client_secret" \
  --access-token "your_access_token"
```

Configuration is saved to `~/.dida365/config.json` with secure permissions.

### 3. Verify Setup

```bash
dida365 auth status
```

## Usage

### Projects

**List all projects:**
```bash
dida365 project list
```

**Get project details:**
```bash
dida365 project get <project-id>
```

### Tasks

**Create a task:**
```bash
dida365 task create \
  --title "Deploy to production" \
  --project-id "proj123" \
  --content "Deploy v2.0.0"
```

**List tasks in a project:**
```bash
dida365 task list proj123
```

**Get task details:**
```bash
dida365 task get task456 --project-id proj123
```

**Update a task:**
```bash
dida365 task update task456 \
  --project-id proj123 \
  --title "New title" \
  --content "Updated content"
```

**Complete a task:**
```bash
dida365 task complete task456 --project-id proj123
```

**Delete a task:**
```bash
dida365 task delete task456 --project-id proj123
```

## Scripting Examples

### Extract specific fields with jq

```bash
# Get task ID from created task
TASK_ID=$(dida365 task create --title "Test" --project-id "proj123" | jq -r .id)

# Get all project names
dida365 project list | jq '.[].name'

# Filter incomplete tasks
dida365 task list proj123 | jq 'map(select(.status == 0))'
```

### Error handling in scripts

```bash
#!/bin/bash
set -e

# Check auth status
if ! dida365 auth status >/dev/null 2>&1; then
  echo "Error: Not authenticated. Run 'dida365 auth configure' first."
  exit 1
fi

# Create task
TASK_JSON=$(dida365 task create --title "Automated task" --project-id "proj123")
if [ $? -ne 0 ]; then
  echo "Failed to create task"
  exit 1
fi

echo "Task created successfully"
echo "$TASK_JSON" | jq .
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Create deployment task
  run: |
    dida365 task create \
      --title "Deployment to ${{ github.ref_name }}" \
      --project-id "${{ secrets.DIDA365_PROJECT_ID }}" \
      --content "SHA: ${{ github.sha }}"
  env:
    DIDA365_CONFIG: ${{ secrets.DIDA365_CONFIG }}
```

## Exit Codes

- `0` - Success
- `1` - Configuration error (missing or invalid config)
- `2` - Authentication error (invalid token)
- `3` - API error (resource not found, bad request, server error)
- `4` - Network error (connection timeout, DNS failure)
- `5` - Client error (invalid arguments, missing flags)

## JSON Output Format

### Success Response

Commands output the resource or array directly:

```json
{
  "id": "task123",
  "projectId": "proj456",
  "title": "Task title",
  "status": 0,
  "sortOrder": 1
}
```

### Error Response

Errors are output to stderr:

```json
{
  "error": "human-readable message",
  "code": "ERROR_CODE"
}
```

## Configuration File

Location: `~/.dida365/config.json`

```json
{
  "client_id": "your_client_id",
  "client_secret": "your_client_secret",
  "access_token": "your_access_token",
  "base_url": "https://dida365.com"
}
```

**Security:** The config file is created with `0600` permissions (user read/write only).

## Development

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/client -v
```

### Building

```bash
go build -o dida365 .
```

### Project Structure

```
dida365-cli/
├── cmd/              # Cobra commands
├── internal/
│   ├── client/       # API client
│   ├── config/       # Config management
│   └── models/       # Data models
├── main.go           # Entry point
└── README.md
```

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `go test ./...`
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Related Projects

- [Dida365 Official Website](https://dida365.com)
- [TickTick (International)](https://ticktick.com)
- [Dida365 API Documentation](https://developer.dida365.com)
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add comprehensive README with usage examples"
```

---

## Task 13: Final Integration Testing

**Files:**
- None (testing only)

**Step 1: Build final binary**

Run:
```bash
go build -o dida365 .
```

**Step 2: Test help commands**

Run:
```bash
./dida365 --help
./dida365 auth --help
./dida365 project --help
./dida365 task --help
```

Expected: All help text displays correctly with subcommands listed

**Step 3: Test version**

Run:
```bash
./dida365 --version
```

Expected: Displays "dida365 version 0.1.0"

**Step 4: Test error cases**

Run:
```bash
./dida365 auth status
./dida365 project list
./dida365 task create --title "Test"
```

Expected: All commands return JSON errors about missing config or flags

**Step 5: Run all tests**

Run:
```bash
go test ./... -v
```

Expected: All tests PASS

**Step 6: Final commit**

```bash
git add .
git commit -m "chore: final integration testing and cleanup"
```

---

## Summary

Implementation complete! The CLI provides:

✅ Full task CRUD operations (create, read, update, delete, complete)
✅ Project read operations (list, get)
✅ Config management with secure file storage
✅ JSON-only output for scripting
✅ Comprehensive error handling with exit codes
✅ Unit tests for all core functionality
✅ Clean architecture (cmd → client → API)

**Next Steps:**
1. Test with real Dida365 API credentials
2. Handle edge cases discovered in real usage
3. Add future enhancements (subtasks, priorities, dates, etc.)

---

## Execution Instructions

Plan complete and saved to `docs/plans/2026-02-20-dida365-cli-implementation.md`.

**Two execution options:**

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach would you like?
