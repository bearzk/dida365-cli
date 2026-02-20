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
