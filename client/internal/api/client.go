package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client communicates with the messenger backend API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates an API client for the given server base URL.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Health checks whether the backend is reachable.
func (c *Client) Health(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("health check failed: %s", resp.Status)
	}

	return string(body), nil
}
