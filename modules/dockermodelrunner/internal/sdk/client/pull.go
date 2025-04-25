package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"

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

	namespace := strings.Split(fullyQualifiedModelName, "/")[0]
	modelName := strings.Split(fullyQualifiedModelName, "/")[1]

	// Verify that the model is pulled successfully, honoring the parent context
	// This is because the pull cancels when the connection is closed.
	err = backoff.RetryNotify(
		func() error {
			if ctx.Err() != nil {
				return backoff.Permanent(ctx.Err())
			}

			model, err := c.InspectModel(ctx, namespace, modelName)
			if err != nil {
				return err
			}
			if model == nil {
				return errors.New("model not found")
			}
			return nil
		},
		backoff.WithContext(backoff.NewExponentialBackOff(), ctx),
		func(err error, _ time.Duration) {
			log.Default().Printf("🙏 Pulling model %s. Please be patient, no progress bar yet! %w", fullyQualifiedModelName, err)
		},
	)
	if err != nil {
		return fmt.Errorf("pull model: %w", err)
	}

	log.Default().Printf("✅ Model %s pulled successfully!", fullyQualifiedModelName)

	return nil
}
