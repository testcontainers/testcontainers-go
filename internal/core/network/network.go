package network

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

type filter string

const (
	idFilter   filter = "id"
	nameFilter filter = "name"
)

// Get returns a network by its ID.
func Get(ctx context.Context, id string) (types.NetworkResource, error) {
	return get(ctx, idFilter, id)
}

// GetByName returns a network by its name.
func GetByName(ctx context.Context, name string) (types.NetworkResource, error) {
	return get(ctx, nameFilter, name)
}

func get(ctx context.Context, f filter, value string) (types.NetworkResource, error) {
	var nw types.NetworkResource // initialize to the zero value

	cli, err := core.NewClient(ctx)
	if err != nil {
		return nw, err
	}
	defer cli.Close()

	filters := filters.NewArgs()
	filters.Add(string(f), value)

	list, err := cli.NetworkList(ctx, types.NetworkListOptions{Filters: filters})
	if err != nil {
		return nw, fmt.Errorf("failed to list networks: %w", err)
	}

	if len(list) == 0 {
		return nw, fmt.Errorf("network %s not found (filtering by %s)", value, f)
	}

	return list[0], nil
}
