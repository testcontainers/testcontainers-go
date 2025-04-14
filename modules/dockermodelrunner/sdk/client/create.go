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
	var bytes []byte

	payload := fmt.Sprintf(`{"from": "%s"}`, fqmn)
	reqURL := c.BaseURL + "/models/create"

	httpClient := http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(payload))
	if err != nil {
		return bytes, fmt.Errorf("new post request (%s): %w", reqURL, err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return bytes, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return bytes, fmt.Errorf("read all: %w", err)
	}
	bytes = body
	return bytes, nil
}
