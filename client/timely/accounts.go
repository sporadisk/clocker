package timely

import (
	"fmt"
	"strconv"
)

func (c *Client) GetCurrentUser() (userId int, err error) {
	err = c.prep()
	if err != nil {
		return 0, fmt.Errorf("c.prep(): %w", err)
	}

	return 0, nil
}

type Account struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *Client) ListAccounts() (accounts []Account, err error) {
	err = c.prep()
	if err != nil {
		return nil, fmt.Errorf("c.prep(): %w", err)
	}

	return nil, nil
}

func (c *Client) SetAccountID(accountID string) error {
	id, err := strconv.ParseInt(accountID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid accountID: %w", err)
	}
	c.accountID = id
	return nil
}
