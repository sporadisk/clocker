package timely

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sporadisk/clocker/console"
	"github.com/sporadisk/clocker/event"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Timely uses different event structures for post and get
type timelyPostEvent struct {
	Day       string    `json:"day"` // format: YYYY-MM-DD
	Hours     int       `json:"hours"`
	Minutes   int       `json:"minutes"`
	Note      string    `json:"note,omitempty"`
	From      time.Time `json:"from,omitzero"`
	To        time.Time `json:"to,omitzero"`
	LabelIDs  []int     `json:"label_ids,omitempty"`
	ProjectID int       `json:"project_id,omitempty"`
}

type timelyGetEvent struct {
	ID      int     `json:"id"`
	Day     string  `json:"day"` // format: YYYY-MM-DD
	User    User    `json:"user"`
	Project Project `json:"project"`
}

func (c *Client) Export(events []*event.Event) error {
	if len(events) == 0 {
		return nil
	}

	timelyEvents := make([]*timelyPostEvent, len(events))
	for i, e := range events {
		te, err := c.eventToTimelyEvent(e)
		if err != nil {
			return fmt.Errorf("eventToTimelyEvent: %w", err)
		}
		timelyEvents[i] = te
	}

	preExistingEvents, err := c.checkForExistingEvents(timelyEvents[0].Day)
	preEventCount := len(preExistingEvents)
	if err != nil {
		return fmt.Errorf("checkForExistingEvents: %w", err)
	}

	if preEventCount > 0 {
		fmt.Println("----------")
		fmt.Printf("Warning: %d events already exist for this user on date %s and project ID %d.\n", preEventCount, timelyEvents[0].Day, c.ProjectID)
		fmt.Println("----------")
		canDelete := console.Confirm("Do you want to delete these events?")
		if !canDelete {
			fmt.Println("Cancelling.")
			return nil
		}

		wait, err := c.DeleteEvents(preExistingEvents)
		if err != nil {
			return fmt.Errorf("DeleteEvents: %w", err)
		}
		fmt.Printf("Deleted %d events.\n", preEventCount)

		sleepSeconds := 2
		if wait {
			// timely gave us a 202 Accepted response, so we should extend the wait
			sleepSeconds = 10
			fmt.Printf("Waiting %d seconds for Timely to process deletions..\n", sleepSeconds)
		}
		time.Sleep(time.Duration(sleepSeconds) * time.Second) // brief pause to ensure Timely processes the deletions
	}

	// the event-batch endpoint has a limit of 100 events per request
	batches := c.makeEventBatches(timelyEvents, 100)
	for i, batch := range batches {
		fmt.Printf("Posting batch %d of %d..\n", i+1, len(batches))
		err := c.PostEventBatch(batch)
		if err != nil {
			return fmt.Errorf("PostEventBatch (batch %d): %w", i, err)
		}
	}
	fmt.Println("Done.")

	return nil
}

func (c *Client) eventToTimelyEvent(e *event.Event) (*timelyPostEvent, error) {
	te := &timelyPostEvent{
		Hours:     e.Hours,
		Minutes:   e.Minutes,
		Note:      eventToNote(e),
		ProjectID: c.ProjectID,
	}

	label, ok := c.tags[e.Category]
	if !ok {
		return te, fmt.Errorf("the specified project has no tag called %q", e.Category)
	}
	te.LabelIDs = []int{label.ID}

	if e.Start.IsZero() || e.End.IsZero() {
		return te, fmt.Errorf("event is missing start or end time")
	}

	te.Day = fmt.Sprintf("%04d-%02d-%02d", e.Date.Year, e.Date.Month, e.Date.Day)
	te.From = e.Start
	te.To = e.End
	return te, nil
}

func eventToNote(e *event.Event) string {
	if e == nil {
		return "nil event"
	}

	// An anglo-centric approach to title-casing.
	// Might add support for more locales later if needed.
	caser := cases.Title(language.English)
	titleCat := caser.String(e.Category)

	if e.Task != "" {
		return fmt.Sprintf("%s: %s", titleCat, e.Task)
	}

	return titleCat
}

func (c *Client) makeEventBatches(events []*timelyPostEvent, batchSize int) [][]*timelyPostEvent {
	var batches [][]*timelyPostEvent
	for i, event := range events {
		batchIndex := i / batchSize
		if batchIndex >= len(batches) {
			batches = append(batches, []*timelyPostEvent{})
		}
		batches[batchIndex] = append(batches[batchIndex], event)
	}

	return batches
}

type eventBatchBody struct {
	Create []*timelyPostEvent `json:"create"`
}

func (c *Client) PostEventBatch(batch []*timelyPostEvent) error {
	err := c.prep()
	if err != nil {
		return fmt.Errorf("c.prep(): %w", err)
	}

	endpoint := fmt.Sprintf("%d/bulk/events", c.AccountID)
	body := &eventBatchBody{
		Create: batch,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	resp, err := c.PostRequest("1.1", endpoint, bodyBytes)
	var batchResp *batchResponse
	if resp != nil {
		var parseErr error
		batchResp, parseErr = parseBatchResponse(resp)
		if parseErr != nil {
			return fmt.Errorf("failed to parse the error response (%w)", parseErr)
		}
	}

	if err != nil {
		if batchResp != nil {
			return fmt.Errorf("Request error (%w) - Response %d: %s", err, resp.Code, batchResp.errorString())
		}
		return fmt.Errorf("c.PostRequest(1.1/%s): %w", endpoint, err)
	}

	if resp == nil {
		return fmt.Errorf("no response received from Timely API")
	}

	if resp.Code == http.StatusOK {
		// Successfully posted the batch
		return nil
	}

	if resp.Code == http.StatusAccepted {
		// The request is being processed asynchronously
		return nil
	}

	return fmt.Errorf("error: resp %d - %s", resp.Code, batchResp.errorString())

}

// ListAllEvents retrieves all events for the specified date (YYYY-MM-DD),
// across all projects.
func (c *Client) ListAllEvents(date string) ([]timelyGetEvent, error) {
	err := c.prep()
	if err != nil {
		return nil, fmt.Errorf("c.prep(): %w", err)
	}
	endpoint := fmt.Sprintf("%d/events", c.AccountID)
	params := map[string]string{
		"day": date,
	}
	resp, err := c.GetRequest("1.1", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("c.GetRequest(1.1/%s): %w", endpoint, err)
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("error: resp %d - %s", resp.Code, string(resp.Body))
	}

	var events []timelyGetEvent
	err = json.Unmarshal(resp.Body, &events)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return events, nil
}

// ListAllEventsForProject retrieves all events for the specified date (YYYY-MM-DD),
// filtered by the client's ProjectID.
func (c *Client) ListAllEventsForProject(date string) ([]timelyGetEvent, error) {
	err := c.prep()
	if err != nil {
		return nil, fmt.Errorf("c.prep(): %w", err)
	}
	endpoint := fmt.Sprintf("%d/projects/%d/events", c.AccountID, c.ProjectID)
	params := map[string]string{
		"day": date,
	}
	resp, err := c.GetRequest("1.1", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("c.GetRequest(1.1/%s): %w", endpoint, err)
	}

	if resp.Code != 200 {
		return nil, fmt.Errorf("error: resp %d - %s", resp.Code, string(resp.Body))
	}
	var events []timelyGetEvent
	err = json.Unmarshal(resp.Body, &events)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return events, nil
}

// checkForExistingEvents checks if this user has already posted events for
// the specified date, and returns true if no events exist.
func (c *Client) checkForExistingEvents(date string) (events []*timelyGetEvent, err error) {
	eventList, err := c.ListAllEventsForProject(date)
	if err != nil {
		return nil, fmt.Errorf("ListAllEventsForProject: %w", err)
	}

	for _, le := range eventList {
		if le.User.ID == c.user.ID {
			events = append(events, &le)
		}
	}

	return events, nil
}

type batchDeleteBody struct {
	Delete []int `json:"delete"`
}

func (c *Client) DeleteEvents(events []*timelyGetEvent) (wait bool, err error) {
	err = c.prep()
	if err != nil {
		return false, fmt.Errorf("c.prep(): %w", err)
	}

	endpoint := fmt.Sprintf("%d/bulk/events", c.AccountID)

	var eventIds []int
	for _, e := range events {
		eventIds = append(eventIds, e.ID)
	}

	body := &batchDeleteBody{
		Delete: eventIds,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return false, fmt.Errorf("json.Marshal: %w", err)
	}

	resp, err := c.PostRequest("1.1", endpoint, bodyBytes)
	if err != nil {
		return false, fmt.Errorf("c.PostRequest(1.1/%s): %w", endpoint, err)
	}

	if resp.Code == http.StatusOK {
		// The request has been processed directly
		return false, nil
	}

	if resp.Code == http.StatusAccepted {
		// The request is being processed asynchronously
		return true, nil // caller should wait a little bit before proceeding
	}

	// some other response was received
	return false, fmt.Errorf("error: resp %d - %s", resp.Code, string(resp.Body))
}
