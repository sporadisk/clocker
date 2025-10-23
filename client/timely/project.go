package timely

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Project struct {
	ID            int            `json:"id"`
	Active        bool           `json:"active"`
	Name          string         `json:"name"`
	Labels        []ProjectLabel `json:"labels,omitempty"`
	RequiredNotes bool           `json:"required_notes"`
}

type ProjectLabel struct {
	ProjectID int `json:"project_id"`
	LabelID   int `json:"label_id"`
}

func (c *Client) GetProjects() error {
	endpoint := fmt.Sprintf("%d/projects", c.AccountID)
	bytes, err := c.GetRequest("1.1", endpoint, nil)

	if err != nil {
		return fmt.Errorf("c.GetRequest(1.1/%s): %w", endpoint, err)
	}

	var projects []Project
	err = json.Unmarshal(bytes.Body, &projects)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	c.projects = projects
	return nil
}

func (c *Client) ListProjects() (string, error) {
	var sb strings.Builder
	for _, project := range c.projects {
		fmt.Fprintf(&sb, " - Project ID: %d, Name: %s\n", project.ID, project.Name)
	}
	return sb.String(), nil
}

func (c *Client) GetProjectByID(id int) (Project, bool) {
	for _, project := range c.projects {
		if project.ID == id {
			return project, true
		}
	}
	return Project{}, false
}
