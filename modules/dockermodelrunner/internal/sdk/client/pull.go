package client

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/testcontainers/testcontainers-go/log"
)

// PullModel creates a model in the Docker Model Runner, by pulling the model from Docker Hub.
func (c *Client) PullModel(ctx context.Context, fullyQualifiedModelName string) error {
	payload := fmt.Sprintf(`{"from": %q}`, fullyQualifiedModelName)
	reqURL := c.baseURL + "/models/create"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("new post request (%s): %w", reqURL, err)
	}

	log.Default().Printf("🙏 Pulling model %s. Please be patient", fullyQualifiedModelName)

	// Check context before making request
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context done before request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	// Check context after getting response
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context done after response: %w", err)
	}

	// The Docker Model Runner returns a 200 status code for a successful pulls
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	done := make(chan error, 1)

	go func() {
		for scanner.Scan() {
			log.Default().Printf(scanner.Text())
		}
		done <- scanner.Err()
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("context done: %w", ctx.Err())
	case err := <-done:
		if err != nil {
			return fmt.Errorf("scan error: %w", err)
		}
	}

	log.Default().Printf("✅ Model %s pulled successfully!", fullyQualifiedModelName)

	return nil
}
