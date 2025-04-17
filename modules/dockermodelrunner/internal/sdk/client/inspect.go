package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner/internal/sdk/types"
)

// InspectModel returns a model by namespace and name
func (c *Client) InspectModel(ctx context.Context, namespace string, name string) (*types.ModelResponse, error) {
	reqURL := c.baseURL + "/models/" + namespace + "/" + name

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new get request (%s): %w", reqURL, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	var model types.ModelResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&model)
	if err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return &model, nil
}
