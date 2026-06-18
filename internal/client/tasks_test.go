package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bearzk/dida365-cli/internal/config"
	"github.com/bearzk/dida365-cli/internal/models"
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
		if req.DueDate != "2026-04-30T15:59:59+0000" {
			t.Errorf("dueDate: got %s, want 2026-04-30T15:59:59+0000", req.DueDate)
		}
		if req.IsAllDay == nil || *req.IsAllDay {
			t.Errorf("isAllDay: got %v, want false", req.IsAllDay)
		}

		dueDate := models.FlexTime{Time: time.Date(2026, 4, 30, 15, 59, 59, 0, time.UTC)}
		isAllDay := false

		task := models.Task{
			ID:        "task456",
			ProjectID: req.ProjectID,
			Title:     req.Title,
			Content:   req.Content,
			DueDate:   &dueDate,
			IsAllDay:  &isAllDay,
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
		DueDate:   "2026-04-30T15:59:59+0000",
		IsAllDay:  &[]bool{false}[0],
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

		dueDate := models.FlexTime{Time: time.Date(2026, 4, 30, 15, 59, 59, 0, time.UTC)}
		isAllDay := false

		task := models.Task{
			ID:        "task456",
			ProjectID: "proj123",
			Title:     "Existing Task",
			DueDate:   &dueDate,
			IsAllDay:  &isAllDay,
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
				{ID: "task1", ProjectID: "proj123", Title: "Task 1", DueDate: &models.FlexTime{Time: time.Date(2026, 4, 30, 15, 59, 59, 0, time.UTC)}, Status: 0, SortOrder: 1},
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

		if req.ID != "task456" {
			t.Errorf("ID: got %s, want task456", req.ID)
		}
		if req.ProjectID != "proj123" {
			t.Errorf("ProjectID: got %s, want proj123", req.ProjectID)
		}
		if req.Title == nil || *req.Title != "Updated Title" {
			t.Error("title not updated correctly")
		}
		if req.DueDate == nil || *req.DueDate != "2026-05-01T00:00:00+0000" {
			t.Errorf("dueDate: got %v, want 2026-05-01T00:00:00+0000", req.DueDate)
		}
		if req.IsAllDay == nil || !*req.IsAllDay {
			t.Errorf("isAllDay: got %v, want true", req.IsAllDay)
		}

		dueDate := models.FlexTime{Time: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)}
		isAllDay := true

		task := models.Task{
			ID:        "task456",
			ProjectID: "proj123",
			Title:     newTitle,
			Content:   newContent,
			DueDate:   &dueDate,
			IsAllDay:  &isAllDay,
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
		DueDate: &[]string{"2026-05-01T00:00:00+0000"}[0],
		IsAllDay: &[]bool{true}[0],
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
