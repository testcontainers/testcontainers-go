package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner/internal/sdk/types"
)

// InspectModel returns a model that is already pulled using the Docker Model Runner format.
// The name of the model is in the format of <name>:<tag>.
// The namespace and name defines Models as OCI Artifacts in Docker Hub, therefore the namespace is the organization and the name is the repository.
// E.g. "ai/smollm2:360M-Q4_K_M". See [Models_as_OCI_Artifacts] for more information.
//
// [Models_as_OCI_Artifacts]: https://hub.docker.com/u/ai
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
