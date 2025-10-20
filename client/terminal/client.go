package terminal

import (
	"fmt"

	"github.com/sporadisk/clocker/format"
)

type Client struct {
	TimeFormat string
}

func (c *Client) Init() error {
	if c.TimeFormat == "" {
		c.TimeFormat = format.TimeHM
	}

	err := format.ValidateTimeFormat(c.TimeFormat)
	if err != nil {
		return fmt.Errorf("ValidateTimeFormat: %w", err)
	}
	return nil
}
