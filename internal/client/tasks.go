package client

import (
	"fmt"

	"github.com/bearzk/dida365-cli/internal/models"
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
	updates.ID = taskID
	updates.ProjectID = projectID
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
