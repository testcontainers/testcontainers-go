package cosmosdb

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

// ContainerPolicy ensures that the request always points to the container. It is required to
// override globalEndpointManager of CosmosDD client, which dynamically updates the [http.Request.Host].
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
