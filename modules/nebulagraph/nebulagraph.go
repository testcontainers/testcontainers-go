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

// RunCluster Run starts Cluster (metad, storaged, graphd) containers within Docker network
func RunCluster(ctx context.Context,
	graphdImg string, graphdCustomizers []testcontainers.ContainerCustomizer,
	storagedImg string, storagedCustomizers []testcontainers.ContainerCustomizer,
	metadImg string, metadCustomizers []testcontainers.ContainerCustomizer,
) (*Cluster, error) {
	// 1. Create a custom network
	netRes, err := network.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	// 2. Start metad
	aggMetadCustomizers := append(defaultMetadContainerCustomizers(netRes), metadCustomizers...)
	metad, err := testcontainers.Run(ctx, metadImg, aggMetadCustomizers...)
	if err != nil {
		errs := []error{fmt.Errorf("failed to start metad: %w", err)}
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes)
		aggErrs := append(errs, errs2...)
		return nil, errors.Join(aggErrs...)
	}

	// 3. Start graphd (needed for storage registration)
	aggGraphdCustomizers := append(defaultGraphdContainerCustomizers(netRes), graphdCustomizers...)
	graphd, err := testcontainers.Run(ctx, graphdImg, aggGraphdCustomizers...)
	if err != nil {
		errs := []error{fmt.Errorf("failed to start graphd: %w", err)}
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes, metad)
		aggErrs := append(errs, errs2...)
		return nil, errors.Join(aggErrs...)
	}

	// 4. Start storaged
	aggStoragedCustomizers := append(defaultStoragedContainerCustomizers(netRes), storagedCustomizers...)
	storaged, err := testcontainers.Run(ctx, storagedImg, aggStoragedCustomizers...)
	if err != nil {
		errs := []error{fmt.Errorf("failed to start storaged: %w", err)}
		fmt.Println("error starting storaged: ", err)
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes, graphd, metad)
		aggErrs := append(errs, errs2...)
		return nil, errors.Join(aggErrs...)
	}

	// 5. Run storage registration command with retry logic
	activator, err := testcontainers.Run(ctx, defaultNebulaConsoleImage, defaultActivatorContainerCustomizers(netRes)...)
	if err != nil {
		errs := []error{fmt.Errorf("failed to start activator container: %w", err)}
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes, storaged, graphd, metad)
		aggErrs := append(errs, errs2...)
		return nil, errors.Join(aggErrs...)
	}

	activatorState, err := activator.State(ctx)
	if !activatorState.Running && activatorState.ExitCode != 0 {
		errs := []error{fmt.Errorf("activator container exited with code %d", activatorState.ExitCode)}
		errs2 := terminateContainersAndRemoveNetwork(ctx, netRes, storaged, graphd, metad)
		aggErrs := append(errs, errs2...)
		return nil, errors.Join(aggErrs...)
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
	host, err := c.graphd.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := c.graphd.MappedPort(ctx, graphdPort)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}

// Terminate stops all NebulaGraph containers
func (c *Cluster) Terminate(ctx context.Context) error {
	errs := terminateContainersAndRemoveNetwork(ctx, c.network, c.graphd, c.metad, c.storaged)
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func terminateContainersAndRemoveNetwork(ctx context.Context, netRes *testcontainers.DockerNetwork, containers ...testcontainers.Container) []error {
	var errs []error
	for _, container := range containers {
		if container != nil {
			if err := container.Terminate(ctx); err != nil {
				errs = append(errs, fmt.Errorf("failed to terminate container: %w", err))
			}
		}
	}

	if err := netRes.Remove(ctx); err != nil {
		errs = append(errs, fmt.Errorf("network remove: %w", err))
	}

	return errs
}
