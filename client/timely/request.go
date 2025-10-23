package timely

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sporadisk/clocker/client"
)

func (c *Client) GetRequest(version, endpoint string, params map[string]string) (*client.Resp, error) {
	endpointUrl := c.timelyEndpoint(version, endpoint)
	req, err := http.NewRequest(http.MethodGet, endpointUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}
	q := url.Values{}
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)

	return c.HttpClient.Do(req)
}

func (c *Client) PostRequest(version, endpoint string, body []byte) (*client.Resp, error) {
	endpointUrl := c.timelyEndpoint(version, endpoint)
	bodyReader := bytes.NewBuffer(body)
	req, err := http.NewRequest(http.MethodPost, endpointUrl, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)

	return c.HttpClient.Do(req)
}

// batchResponse is a struct that is returned from all of Timely's batch
// endpoints.
type batchResponse struct {
	DeletedIDs []int               `json:"deleted_ids"`
	CreatedIDs []int               `json:"created_ids"`
	UpdatedIDs []int               `json:"updated_ids"`
	Errors     map[string][]string `json:"errors"`
	// There's also a "job" field, but its type or structure is undocumented. All the example responses have it set to null.
}

func (br *batchResponse) errorString() string {
	sb := strings.Builder{}
	for field, errs := range br.Errors {
		for i, err := range errs {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "[%s: %s]", field, err)
		}
	}
	return sb.String()
}

func parseBatchResponse(resp *client.Resp) (*batchResponse, error) {
	var br batchResponse
	err := json.Unmarshal(resp.Body, &br)
	if err != nil {
		// Print the response body for debugging
		fmt.Printf("resp %d: %q", resp.Code, string(resp.Body))
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return &br, nil
}
