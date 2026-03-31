package network

import (
	"context"
	"fmt"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

const (
	// FilterByID uses to filter network by identifier.
	FilterByID = "id"

	// FilterByName uses to filter network by name.
	FilterByName = "name"
)

// Get returns a network by its ID.
func Get(ctx context.Context, id string) (network.Inspect, error) {
	return get(ctx, FilterByID, id)
}

// GetByName returns a network by its name.
func GetByName(ctx context.Context, name string) (network.Inspect, error) {
	return get(ctx, FilterByName, name)
}

func get(ctx context.Context, filter string, value string) (network.Inspect, error) {
	var nw network.Inspect // initialize to the zero value

	cli, err := core.NewClient(ctx)
	if err != nil {
		return nw, err
	}
	defer cli.Close()

	result, err := cli.NetworkList(ctx, client.NetworkListOptions{
		Filters: make(client.Filters).Add(filter, value),
	})
	if err != nil {
		return nw, fmt.Errorf("failed to list networks: %w", err)
	}

	if len(result.Items) == 0 {
		return nw, fmt.Errorf("network %s not found (filtering by %s)", value, filter)
	}

	// Get the full inspect result for the found network
	inspectResult, err := cli.NetworkInspect(ctx, result.Items[0].ID, client.NetworkInspectOptions{})
	if err != nil {
		return nw, fmt.Errorf("failed to inspect network: %w", err)
	}

	return inspectResult.Network, nil
}
