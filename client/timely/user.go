package timely

import (
	"encoding/json"
	"fmt"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (c *Client) GetCurrentUser() (User, error) {
	var user User
	err := c.prep()
	if err != nil {
		return User{}, err
	}
	endpoint := fmt.Sprintf("%d/users/current", c.AccountID)
	bytes, err := c.GetRequest("1.1", endpoint, nil)
	if err != nil {
		return user, fmt.Errorf("c.GetRequest(1.1/%s): %w", endpoint, err)
	}

	err = json.Unmarshal(bytes.Body, &user)
	if err != nil {
		return user, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return user, nil
}
