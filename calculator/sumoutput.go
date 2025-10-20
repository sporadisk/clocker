package calculator

import (
	"fmt"

	"github.com/sporadisk/clocker/client/terminal"
	"github.com/sporadisk/clocker/format"
)

func (c *Calculator) LoadSummaryOutput() error {
	// Currently only terminal output is supported
	return c.LoadTerminalOutput()
}

func (c *Calculator) LoadTerminalOutput() error {
	defaultTimeFormat := format.TimeHM
	if c.Conf.Output != nil && c.Conf.Output.Params != nil {
		format, ok := c.Conf.Output.Params["timeFormat"]
		if ok {
			defaultTimeFormat = format
		}
	}

	// currently only one output type: Terminal
	termClient := &terminal.Client{
		TimeFormat: defaultTimeFormat,
	}
	err := termClient.Init()
	if err != nil {
		return fmt.Errorf("terminal.Client.Init: %w", err)
	}

	c.SummaryOutput = termClient
	return nil
}
