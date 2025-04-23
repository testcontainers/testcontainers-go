package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// PullModel creates a model in the Docker Model Runner, by pulling the model from Docker Hub.
func (c *Client) PullModel(ctx context.Context, fqmn string) ([]byte, error) {
	payload := fmt.Sprintf(`{"from": %q}`, fqmn)
	reqURL := c.baseURL + "/models/create"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("new post request (%s): %w", reqURL, err)
	}

	log.Default().Printf("🙏 Pulling model %s. Please be patient, no progress bar yet!", fqmn)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	// The Docker Model Runner returns a 200 status code for a successful pulls
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read all: %w", err)
	}

	log.Default().Printf("✅ Model %s pulled successfully!", fqmn)

	return body, nil
}
