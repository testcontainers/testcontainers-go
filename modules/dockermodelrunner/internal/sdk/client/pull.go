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

	logger := log.Default()

	logger.Printf("🙏 Pulling model %s. Please be patient", fullyQualifiedModelName)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	// The Docker Model Runner returns a 200 status code for a successful pulls
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	// TODO: use a progressbar instead of multiple line output.
	for scanner.Scan() {
		logger.Printf(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan error: %w", err)
	}

	logger.Printf("✅ Model %s pulled successfully!", fullyQualifiedModelName)

	return nil
}
