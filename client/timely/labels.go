package timely

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Label struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Sequence int     `json:"sequence"`
	ParentID *int    `json:"parent_id,omitempty"`
	Active   bool    `json:"active"`
	Children []Label `json:"children,omitempty"`
}

func (c *Client) GetProjectLabels() (labels []Label, err error) {
	err = c.GetAllLabels()
	if err != nil {
		return nil, fmt.Errorf("GetAllLabels: %w", err)
	}

	label, ok := findLabelByID(c.labels, c.ProjectID)
	if !ok {
		return nil, fmt.Errorf("project label with ID %d not found", c.ProjectID)
	}

	labels = label.Children
	return labels, nil
}

func findLabelByID(labels []Label, id int) (label Label, ok bool) {
	for _, label := range labels {
		if label.ID == id {
			return label, true
		}
		if len(label.Children) > 0 {
			childLabel, ok := findLabelByID(label.Children, id)
			if ok {
				return childLabel, true
			}
		}
	}
	return Label{}, false
}

func (c *Client) GetAllLabels() (err error) {
	if c.labels != nil {
		return nil // labels are already loaded
	}

	err = c.prep()
	if err != nil {
		return fmt.Errorf("c.prep(): %w", err)
	}

	if c.AccountID == 0 {
		return fmt.Errorf("account ID is not set")
	}

	endpoint := fmt.Sprintf("%d/labels", c.AccountID)
	bytes, err := c.GetRequest("1.1", endpoint, nil)
	if err != nil {
		return fmt.Errorf("c.GetRequest(1.1/%s): %w", endpoint, err)
	}

	var labels []Label
	err = json.Unmarshal(bytes.Body, &labels)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	c.labels = labels
	return nil
}

func (c *Client) ListRootLabels() (string, error) {
	var sb strings.Builder
	for _, label := range c.labels {
		if label.Active && label.ParentID == nil {
			_, err := sb.WriteString(fmt.Sprintf(" - ID: %d  Name: %s\n", label.ID, label.Name))
			if err != nil {
				return "", fmt.Errorf("writing to string builder: %w", err)
			}
		}
	}

	return sb.String(), nil
}

func (c *Client) mapTags() error {
	c.tags = make(map[string]Label)

	project, ok := c.GetProjectByID(c.ProjectID)
	if !ok {
		return fmt.Errorf("could not find the project")
	}

	for _, pl := range project.Labels {
		label, ok := findLabelByID(c.labels, pl.LabelID)
		if !ok {
			return fmt.Errorf("could not find label with ID %d", pl.LabelID)
		}
		if label.ParentID == nil {
			continue // this is a root label
		}

		tagLower := strings.ToLower(label.Name)
		c.tags[tagLower] = label
	}

	return nil
}
