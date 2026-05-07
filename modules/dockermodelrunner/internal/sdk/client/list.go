package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner/internal/sdk/types"
)

// ListModels lists all models that are already pulled using the Docker Model Runner format.
func (c *Client) ListModels(ctx context.Context) ([]types.ModelResponse, error) {
	reqURL := c.baseURL + "/models"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new get request (%s): %w", reqURL, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	var models []types.ModelResponse
	err = decoder.Decode(&models)
	if err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return models, nil
}
