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
