package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CreateModel creates a model
func (c *Client) CreateModel(ctx context.Context, fqmn string) ([]byte, error) {
	payload := fmt.Sprintf(`{"from": %q}`, fqmn)
	reqURL := c.baseURL + "/models/create"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("new post request (%s): %w", reqURL, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	// The Docker Model Runner returns a 200 status code for a successful pull
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read all: %w", err)
	}

	return body, nil
}
