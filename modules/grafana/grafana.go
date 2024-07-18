package grafana

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

// GrafanaContainer represents the Grafana container type used in the module
type GrafanaContainer struct {
	testcontainers.Container
}

// Run creates an instance of the Grafana container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*GrafanaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &GrafanaContainer{Container: container}, nil
}
