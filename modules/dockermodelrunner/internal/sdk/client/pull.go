package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// PullModel creates a model in the Docker Model Runner, by pulling the model from Docker Hub.
func (c *Client) PullModel(ctx context.Context, fullyQualifiedModelName string) error {
	payload := fmt.Sprintf(`{"from": %q}`, fullyQualifiedModelName)
	reqURL := c.baseURL + "/models/create"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("new post request (%s): %w", reqURL, err)
	}

	log.Default().Printf("🙏 Pulling model %s. Please be patient, no progress bar yet!", fullyQualifiedModelName)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	// The Docker Model Runner returns a 200 status code for a successful pulls
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Default().Printf("✅ Model %s pulled successfully!", fullyQualifiedModelName)

	return nil
}
