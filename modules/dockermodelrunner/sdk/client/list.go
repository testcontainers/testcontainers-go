package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models, fmt.Errorf("read all: %w", err)
	}

	err = json.Unmarshal(body, &models)
	if err != nil {
		return models, fmt.Errorf("json unmarshal: %w", err)
	}

	return models, nil
}
