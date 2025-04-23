package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// pullOptions contains options for pulling a model
type pullOptions struct {
	// Progress is an optional writer to show download progress
	Progress ProgressWriter
}

// PullOption is a function that configures PullOptions
type PullOption func(*pullOptions)

// PullModel creates a model in the Docker Model Runner, by pulling the model from Docker Hub.
func (c *Client) PullModel(ctx context.Context, fqmn string, opts ...PullOption) error {
	options := &pullOptions{}
	for _, opt := range opts {
		opt(options)
	}

	payload := fmt.Sprintf(`{"from": %q}`, fqmn)
	reqURL := c.baseURL + "/models/create"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("new post request (%s): %w", reqURL, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	// The Docker Model Runner returns a 200 status code for a successful pulls
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// If progress writer is provided, create a progress reader
	var reader io.Reader = resp.Body
	if options.Progress != nil {
		// If we have Content-Length, use it
		if resp.ContentLength > 0 {
			options.Progress.SetTotal(resp.ContentLength)
		} else {
			// Otherwise use indeterminate progress
			options.Progress.SetTotal(-1)
		}
		reader = io.TeeReader(resp.Body, options.Progress)
	}

	// Read the response
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("copy response: %w", err)
	}

	log.Default().Printf("✅ Model %s pulled successfully!", fqmn)

	return nil
}
