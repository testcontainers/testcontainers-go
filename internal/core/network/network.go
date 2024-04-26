package network

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// Get returns a network by its ID.
func Get(ctx context.Context, id string) (types.NetworkResource, error) {
	var nw types.NetworkResource // initialize to the zero value

	cli, err := core.NewClient(ctx)
	if err != nil {
		return nw, err
	}
	defer cli.Close()

	filters := filters.NewArgs()
	filters.Add("id", id)

	list, err := cli.NetworkList(ctx, types.NetworkListOptions{Filters: filters})
	if err != nil {
		return nw, fmt.Errorf("failed to list networks: %w", err)
	}

	if len(list) == 0 {
		return nw, fmt.Errorf("network %s not found", id)
	}

	return list[0], nil
}
