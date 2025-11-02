package client

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// HttpClient implements a client for sending HTTP requests, while automating
// some busywork such as closing response bodies.
type HttpClient struct {
	Client http.Client
}

func NewHttpClient(timeout time.Duration) *HttpClient {
	hc := &HttpClient{
		Client: http.Client{
			Timeout: timeout,
		},
	}
	return hc
}

type Resp struct {
	Code   int
	Body   []byte
	Header http.Header
}

// Do is a simplified version of http.Client.Do that reads the response body
// and returns it as a byte slice.
func (hc *HttpClient) Do(req *http.Request) (*Resp, error) {
	hr, err := hc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client error: %w", err)
	}

	defer hr.Body.Close()
	body, err := io.ReadAll(hr.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return &Resp{
		Code:   hr.StatusCode,
		Body:   body,
		Header: hr.Header,
	}, nil
}
