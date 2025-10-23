package calculator

import (
	"fmt"
	"os"
	"time"

	"github.com/sporadisk/clocker/config"
	"github.com/sporadisk/clocker/console"
	"github.com/sporadisk/clocker/event"
	"github.com/sporadisk/clocker/logentry"
	"github.com/sporadisk/clocker/parameter"
	"github.com/sporadisk/clocker/summary"
)

type Calculator struct {
	Conf              *config.Config
	EventExporter     event.Exporter
	Subscriber        logentry.Subscriber
	SummaryOutput     summary.Output
	DefaultFullDay    time.Duration
	CategoryParseMode string

	eventInbox chan inboxEvent
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

	if c.Conf.Calc != nil && c.Conf.Calc.CategoryParseMode != "" {
		parseMode, err := parameter.Validate(c.Conf.Calc.CategoryParseMode, []string{"v1", "v2"})
		if err != nil {
			return fmt.Errorf("validation failure for category parse mode: %w", err)
		}
		c.CategoryParseMode = parseMode
	}

	// Start the inbox processing goroutine
	c.eventInbox = make(chan inboxEvent, 100)
	go func() {
		err := c.WaitForEntries()
		if err != nil {
			fmt.Printf("WaitForEntries: %s\n", err.Error())
			os.Exit(1)
		}
	}()

	// Subscribe to log entries
	err = c.Subscriber.Subscribe(c)
	if err != nil {
		return fmt.Errorf("Subscriber.Subscribe: %w", err)
	}

	// Start the event processing loop
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

func (c *Calculator) AskAndExport(summaryEvents []*event.Event) error {
	if !console.Confirm(fmt.Sprintf("Export log events to %s?", c.Conf.Exporter.Name)) {
		fmt.Println("Export denied.")
		return nil
	}

	fmt.Println("Export started.")
	err := c.EventExporter.Export(summaryEvents)
	if err != nil {
		return fmt.Errorf("EventExporter.Export: %w", err)
	}
	fmt.Println("Export completed.")

	return nil
}
