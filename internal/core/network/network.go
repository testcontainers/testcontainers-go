package network

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

const (
	// FilterByID uses to filter network by identifier.
	FilterByID = "id"

	// FilterByName uses to filter network by name.
	FilterByName = "name"

	Bridge        string = "bridge"         // Bridge network name (as well as driver)
	ReaperDefault string = "reaper_default" // Default network name when bridge is not available
)

// Get returns a network by its ID.
func Get(ctx context.Context, id string) (types.NetworkResource, error) {
	return get(ctx, FilterByID, id)
}

// GetByName returns a network by its name.
func GetByName(ctx context.Context, name string) (types.NetworkResource, error) {
	return get(ctx, FilterByName, name)
}

func get(ctx context.Context, filter string, value string) (types.NetworkResource, error) {
	var nw types.NetworkResource // initialize to the zero value

	cli, err := core.NewClient(ctx)
	if err != nil {
		return nw, err
	}
	defer cli.Close()

	filters := filters.NewArgs()
	filters.Add(filter, value)

	list, err := cli.NetworkList(ctx, types.NetworkListOptions{Filters: filters})
	if err != nil {
		return nw, fmt.Errorf("failed to list networks: %w", err)
	}

	if len(list) == 0 {
		return nw, fmt.Errorf("network %s not found (filtering by %s)", value, filter)
	}

	return list[0], nil
}

func GetGatewayIP(ctx context.Context) (string, error) {
	defaultNetwork, err := GetDefault(ctx)
	if err != nil {
		return "", err
	}

	nw, err := Get(ctx, defaultNetwork)
	if err != nil {
		return "", err
	}

	var ip string
	for _, config := range nw.IPAM.Config {
		if config.Gateway != "" {
			ip = config.Gateway
			break
		}
	}
	if ip == "" {
		return "", errors.New("failed to get gateway IP from network settings")
	}

	return ip, nil
}

func GetDefault(ctx context.Context) (string, error) {
	cli, err := core.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()

	// Get list of available networks
	networkResources, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return "", err
	}

	reaperNetwork := ReaperDefault

	reaperNetworkExists := false

	for _, net := range networkResources {
		if net.Name == Bridge {
			return net.Name, nil
		}

		if net.Name == reaperNetwork {
			reaperNetworkExists = true
		}
	}

	// Create a bridge network for the container communications
	if !reaperNetworkExists {
		_, err = cli.NetworkCreate(ctx, reaperNetwork, types.NetworkCreate{
			Driver:     Bridge,
			Attachable: true,
			Labels:     core.DefaultLabels(core.SessionID()),
		})
		if err != nil {
			return "", err
		}
	}

	return reaperNetwork, nil
}
