package client

import (
	"fmt"

	"github.com/bearzk/dida365-cli/internal/models"
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

// GetProjectData returns the full typed response from the project data endpoint,
// including the project details, tasks, and columns.
// For normal task listing use ListTasks instead.
func (c *Client) GetProjectData(projectID string) (*models.ProjectData, error) {
	path := fmt.Sprintf("/open/v1/project/%s/data", projectID)
	var data models.ProjectData
	if err := c.doRequest("GET", path, nil, &data); err != nil {
		return nil, fmt.Errorf("failed to get project data: %w", err)
	}
	return &data, nil
}
