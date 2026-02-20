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
