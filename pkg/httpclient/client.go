// Package httpclient provides an HTTP client for executing outbound API calls.
package httpclient

import (
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

// Client wraps net/http.Client for executing outbound HTTP requests.
type Client struct {
	c *http.Client
}

// New creates a Client with the given timeout.
// If timeout is 0, the default 30-second timeout is used.
func New(timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &Client{c: &http.Client{Timeout: timeout}}
}

// Do executes the given HTTP request and returns the response.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.c.Do(req)
}
