package network

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"

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

	list, err := cli.NetworkList(ctx, network.ListOptions{
		Filters: filters.NewArgs(filters.Arg(filter, value)),
	})
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

	nw, err := GetByName(ctx, defaultNetwork)
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
	networkResources, err := cli.NetworkList(ctx, network.ListOptions{})
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
		_, err = cli.NetworkCreate(ctx, reaperNetwork, network.CreateOptions{
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

// New creates a new network.
func New(ctx context.Context, req Request) (network.CreateResponse, error) {
	if req.Labels == nil {
		req.Labels = make(map[string]string)
	}

	// add the labels that the reaper will use to terminate the network to the request
	for k, v := range core.DefaultLabels(core.SessionID()) {
		req.Labels[k] = v
	}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return network.CreateResponse{}, err
	}
	defer cli.Close()

	nc := network.CreateOptions{
		Driver:     req.Driver,
		Internal:   req.Internal,
		EnableIPv6: req.EnableIPv6,
		Attachable: req.Attachable,
		Labels:     req.Labels,
		IPAM:       req.IPAM,
	}

	return cli.NetworkCreate(ctx, req.Name, nc)
}
