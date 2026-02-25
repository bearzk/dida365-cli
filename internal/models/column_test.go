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
