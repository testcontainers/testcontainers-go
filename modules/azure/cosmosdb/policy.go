package cosmosdb

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

// ContainerPolicy ensures that requests always target the CosmosDB emulator container endpoint.
// It overrides the CosmosDB client's globalEndpointManager, which would otherwise dynamically
// update [http.Request.Host] based on global endpoint discovery, pinning all requests to the container.
type ContainerPolicy struct {
	endpoint string
}

func NewContainerPolicy(ctx context.Context, c *Container) (*ContainerPolicy, error) {
	endpoint, err := c.PortEndpoint(ctx, defaultPort, "")
	if err != nil {
		return nil, fmt.Errorf("port endpoint: %v", err)
	}

	return &ContainerPolicy{
		endpoint: endpoint,
	}, nil
}

func (p *ContainerPolicy) Do(req *policy.Request) (*http.Response, error) {
	req.Raw().Host = p.endpoint
	req.Raw().URL.Host = p.endpoint

	return req.Next()
}

// ClientOptions returns Azure CosmosDB client options that contain ContainerPolicy.
func (p *ContainerPolicy) ClientOptions() *azcosmos.ClientOptions {
	return &azcosmos.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			PerRetryPolicies: []policy.Policy{p},
		},
	}
}
