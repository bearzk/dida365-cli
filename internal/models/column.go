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
