package calculator

import (
	"fmt"
	"time"

	"github.com/sporadisk/clocker/config"
	"github.com/sporadisk/clocker/event"
	"github.com/sporadisk/clocker/logentry"
	"github.com/sporadisk/clocker/summary"
)

type Calculator struct {
	Conf           *config.Config
	EventExporter  event.Exporter
	Subscriber     logentry.Subscriber
	SummaryOutput  summary.Output
	DefaultFullDay time.Duration
}

func (c *Calculator) Start() error {
	err := c.getDefaultFullDay()
	if err != nil {
		return fmt.Errorf("c.getDefaultFullDay: %w", err)
	}

	// Initialize the exporter, if one has been configured
	if c.Conf.Exporter != nil {
		exporter, err := LoadExporter(c.Conf.Exporter)
		if err != nil {
			return fmt.Errorf("Error loading exporter: %w", err)
		}

		c.EventExporter = exporter
	}

	err = c.LoadSummaryOutput()
	if err != nil {
		return fmt.Errorf("LoadSummaryOutput: %w", err)
	}

	err = c.Subscriber.Subscribe(c)
	if err != nil {
		return fmt.Errorf("Subscriber.Subscribe: %w", err)
	}
	return nil
}

func (c *Calculator) Receive(entries []logentry.Entry) error {
	summary := LogSum(entries, c.DefaultFullDay)
	err := c.SummaryOutput.OutputSummary(summary)
	if err != nil {
		return fmt.Errorf("SummaryOutput.Output: %w", err)
	}

	if c.EventExporter != nil {
		// TODO: Generate exportable events during summary calculation
	}
	return nil
}

func (c *Calculator) getDefaultFullDay() error {
	if c.Conf.DefaulltFullDay != "" {
		dfd, err := time.ParseDuration(c.Conf.DefaulltFullDay)
		if err != nil {
			return fmt.Errorf("parsing default full day duration: %w", err)
		}
		c.DefaultFullDay = dfd
		return nil
	}

	// Fallback to 7.5 hours
	c.DefaultFullDay = 450 * time.Minute
	return nil
}
