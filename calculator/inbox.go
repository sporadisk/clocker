package calculator

import (
	"fmt"
	"time"

	"github.com/sporadisk/clocker/logentry"
)

type inboxEvent struct {
	timestamp time.Time
	entries   []logentry.Entry
}

func (c *Calculator) Receive(entries []logentry.Entry) error {
	c.eventInbox <- inboxEvent{
		timestamp: time.Now(),
		entries:   entries,
	}

	return nil
}

func (c *Calculator) WaitForEntries() error {
	for {
		inboxItem := <-c.eventInbox
		err := c.Process(inboxItem.entries)
		if err != nil {
			return fmt.Errorf("c.Process: %w", err)
		}

		// Todo: Add graceful shutdown handling
	}
}

func (c *Calculator) Process(entries []logentry.Entry) error {
	summary := &LogSummary{
		Entries:      entries,
		FullDay:      c.DefaultFullDay,
		CatParseMode: c.CategoryParseMode,
	}

	summaryResult, summaryEvents := summary.Sum()

	err := c.SummaryOutput.OutputSummary(summaryResult)
	if err != nil {
		return fmt.Errorf("SummaryOutput.Output: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	// don't try to export if the summary is for today
	// (as a rule, the log for today is probably incomplete)
	canExport := (today != summaryResult.Date.String())

	if c.EventExporter != nil && canExport {
		err := c.AskAndExport(summaryEvents)
		if err != nil {
			return fmt.Errorf("AskAndExport: %w", err)
		}
	}

	return nil
}
