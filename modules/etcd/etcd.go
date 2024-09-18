package etcd

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

// etcdContainer represents the etcd container type used in the module
type etcdContainer struct {
	testcontainers.Container
}

// Run creates an instance of the etcd container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*etcdContainer, error) {
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
	var c *etcdContainer
	if container != nil {
		c = &etcdContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
