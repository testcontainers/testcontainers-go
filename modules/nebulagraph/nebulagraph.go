package nebulagraph

import (
	"context"
	"errors"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

// Cluster represents a running NebulaGraph cluster for testing
type Cluster struct {
	graphd   testcontainers.Container
	metad    testcontainers.Container
	storaged testcontainers.Container
	network  *testcontainers.DockerNetwork
}

// RunCluster starts a NebulaGraph cluster (metad, storaged, graphd and activator) containers within a Docker network
func RunCluster(ctx context.Context,
	graphdImg string, graphdCustomizers []testcontainers.ContainerCustomizer,
	storagedImg string, storagedCustomizers []testcontainers.ContainerCustomizer,
	metadImg string, metadCustomizers []testcontainers.ContainerCustomizer,
) (*Cluster, error) {
	// 1. Create a custom network
	netRes, err := network.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("new nebulagraph network: %w", err)
	}

	// 2. Start metad
	aggMetadCustomizers := append(defaultMetadContainerCustomizers(netRes), metadCustomizers...)
	metad, err := testcontainers.Run(ctx, metadImg, aggMetadCustomizers...)
	if err != nil {
		errs := []error{fmt.Errorf("run metad container: %w", err)}
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes)
		errs = append(errs, errs2...)
		return nil, errors.Join(errs...)
	}

	// 3. Start graphd (needed for storage registration)
	aggGraphdCustomizers := append(defaultGraphdContainerCustomizers(netRes), graphdCustomizers...)
	graphd, err := testcontainers.Run(ctx, graphdImg, aggGraphdCustomizers...)
	if err != nil {
		errs := []error{fmt.Errorf("run graphd container: %w", err)}
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes, metad)
		errs = append(errs, errs2...)
		return nil, errors.Join(errs...)
	}

	// 4. Start storaged
	aggStoragedCustomizers := append(defaultStoragedContainerCustomizers(netRes), storagedCustomizers...)
	storaged, err := testcontainers.Run(ctx, storagedImg, aggStoragedCustomizers...)
	if err != nil {
		errs := []error{fmt.Errorf("run storaged container: %w", err)}
		fmt.Println("error starting storaged: ", err)
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes, graphd, metad)
		errs = append(errs, errs2...)
		return nil, errors.Join(errs...)
	}

	// 5. Run storage registration command with retry logic
	activator, err := testcontainers.Run(ctx, defaultNebulaConsoleImage, defaultActivatorContainerCustomizers(netRes)...)
	if err != nil {
		errs := []error{fmt.Errorf("run activator container: %w", err)}
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes, storaged, graphd, metad)
		errs = append(errs, errs2...)
		return nil, errors.Join(errs...)
	}

	activatorState, err := activator.State(ctx)
	if err != nil {
		errs := []error{fmt.Errorf("get activator container state: %w", err)}
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes, storaged, graphd, metad, activator)
		errs = append(errs, errs2...)
		return nil, errors.Join(errs...)
	}

	if !activatorState.Running && activatorState.ExitCode != 0 {
		errs := []error{fmt.Errorf("activator container not running or exited with code %d", activatorState.ExitCode)}
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes, storaged, graphd, metad)
		errs = append(errs, errs2...)
		return nil, errors.Join(errs...)
	}

	return &Cluster{
		graphd:   graphd,
		metad:    metad,
		storaged: storaged,
		network:  netRes,
	}, nil
}

// ConnectionString returns the host:port for connecting to NebulaGraph graphd
func (c *Cluster) ConnectionString(ctx context.Context) (string, error) {
	return c.graphd.PortEndpoint(ctx, graphdPort, "")
}

// Terminate stops all NebulaGraph containers
func (c *Cluster) Terminate(ctx context.Context) error {
	errs := terminateContainersAndRemoveNetwork(ctx, c.network, c.graphd, c.metad, c.storaged)
	return errors.Join(errs...)
}

func terminateContainersAndRemoveNetwork(ctx context.Context, netRes *testcontainers.DockerNetwork, containers ...testcontainers.Container) []error {
	var errs []error
	for _, ctr := range containers {
		if ctr != nil {
			if err := ctr.Terminate(ctx); err != nil {
				errs = append(errs, fmt.Errorf("terminate container: %w", err))
			}
		}
	}

	if err := netRes.Remove(ctx); err != nil {
		errs = append(errs, fmt.Errorf("network remove: %w", err))
	}

	return errs
}
