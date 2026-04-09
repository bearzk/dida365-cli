package models

// Task represents a Dida365 task
type Task struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"projectId"`
	Title         string     `json:"title"`
	Content       string     `json:"content,omitempty"`
	StartDate     *FlexTime  `json:"startDate,omitempty"`
	DueDate       *FlexTime  `json:"dueDate,omitempty"`
	TimeZone      string     `json:"timeZone,omitempty"`
	IsAllDay      *bool      `json:"isAllDay,omitempty"`
	Status        int        `json:"status"`             // 0=normal, 2=completed
	Priority      int        `json:"priority,omitempty"` // 0=none, 1=low, 3=med, 5=high
	CompletedTime *FlexTime  `json:"completedTime,omitempty"`
	SortOrder     int        `json:"sortOrder"`
	ColumnID      string     `json:"columnId,omitempty"`
}

// TaskCreate represents the payload for creating a new task
type TaskCreate struct {
	Title     string `json:"title"`
	ProjectID string `json:"projectId"`
	Content   string `json:"content,omitempty"`
	DueDate   string `json:"dueDate,omitempty"`
	IsAllDay  *bool  `json:"isAllDay,omitempty"`
}

// TaskUpdate represents the payload for updating a task
type TaskUpdate struct {
	ID        string  `json:"id"`
	ProjectID string  `json:"projectId"`
	Title     *string `json:"title,omitempty"`
	Content   *string `json:"content,omitempty"`
	ColumnID  *string `json:"columnId,omitempty"`
	DueDate   *string `json:"dueDate,omitempty"`
	IsAllDay  *bool   `json:"isAllDay,omitempty"`
}
