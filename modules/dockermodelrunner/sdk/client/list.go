package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner/sdk/types"
)

// ListModels lists all models
func (c *Client) ListModels(ctx context.Context) ([]types.ModelResponse, error) {
	var models []types.ModelResponse

	reqURL := c.baseURL + "/models"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return models, fmt.Errorf("new get request (%s): %w", reqURL, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return models, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&models)
	if err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return models, nil
}
