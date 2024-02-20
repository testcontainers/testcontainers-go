package opensearch

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
	"github.com/testcontainers/testcontainers-go"
)

// OpenSearchContainer represents the OpenSearch container type used in the module
type OpenSearchContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the OpenSearch container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OpenSearchContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "opensearchproject/opensearch:2.11.1",
		ExposedPorts: []string{"9200/tcp", "9600/tcp"},
		Env: map[string]string{
			"discovery.type": "single-node",
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Ulimits = []*units.Ulimit{
				{
					Name: "memlock",
					Soft: -1, // Set memlock to unlimited (no soft or hard limit)
					Hard: -1,
				},
				{
					Name: "nofile",
					Soft: 65536, // Maximum number of open files for the opensearch user - set to at least 65536
					Hard: 65536,
				},
			}
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &OpenSearchContainer{Container: container}, nil
}
