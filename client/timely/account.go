package timely

import (
	"encoding/json"
	"fmt"
)

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

	bytes, err := c.GetRequest("1.1", "accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("c.GetRequest(1.1/accounts): %w", err)
	}

	resp := []Account{}
	err = json.Unmarshal(bytes.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return resp, nil
}

func (c *Client) SelectAccount() error {
	inputAccountID := c.AccountID

	accounts, err := c.ListAccounts()
	if err != nil {
		return fmt.Errorf("ListAccounts: %w", err)
	}

	if len(accounts) <= 0 {
		return fmt.Errorf("no accounts available for this user")
	}

	if len(accounts) == 1 {
		if inputAccountID == 0 {
			c.AccountID = accounts[0].ID
			return nil
		}

		// inputAccountID != 0
		if accounts[0].ID != inputAccountID {
			return fmt.Errorf("specified accountId %d does not match the only available accountId %d", inputAccountID, accounts[0].ID)
		}
		c.AccountID = accounts[0].ID
		return nil
	}

	// len(accounts) > 1
	if inputAccountID == 0 {
		for _, acct := range accounts {
			fmt.Printf("Account ID: %d, Name: %s\n", acct.ID, acct.Name)
		}
		return fmt.Errorf("multiple accounts available: please specify accountId in the exporter parameters")
	}

	// len(accounts) > 1 && inputAccountID != 0
	for _, acct := range accounts {
		if acct.ID == inputAccountID {
			c.AccountID = acct.ID
			return nil
		}
	}

	// inputAccountID not found
	for _, acct := range accounts {
		fmt.Println("Available accounts:")
		fmt.Printf("Account ID: %d, Name: %s\n", acct.ID, acct.Name)
	}
	return fmt.Errorf("specified accountId %d not found among available accounts", inputAccountID)
}
