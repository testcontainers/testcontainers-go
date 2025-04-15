package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner/sdk/types"
)

// InspectModel returns a model by namespace and name
func (c *Client) InspectModel(ctx context.Context, namespace string, name string) (types.ModelResponse, error) {
	var model types.ModelResponse

	httpClient := http.Client{}

	reqURL := c.baseURL + "/models/" + namespace + "/" + name

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return model, fmt.Errorf("new get request (%s): %w", reqURL, err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return model, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model, fmt.Errorf("read all: %w", err)
	}

	err = json.Unmarshal(body, &model)
	if err != nil {
		return model, fmt.Errorf("json unmarshal: %w", err)
	}

	return model, nil
}
