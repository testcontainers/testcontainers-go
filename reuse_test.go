package testcontainers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

func TestGenericContainer_stop_start_withReuse(t *testing.T) {
	containerName := "my-nginx"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        nginxAlpineImage,
			ExposedPorts: []string{"8080/tcp"},
			Name:         containerName,
		},
		Reuse:   true,
		Started: true,
	}

	ctr, err := testcontainers.GenericContainer(context.Background(), req)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.NotNil(t, ctr)

	err = ctr.Stop(context.Background(), nil)
	require.NoError(t, err)

	// Run another container with same container name:
	// The checks for the exposed ports must not fail when restarting the container.
	ctr1, err := testcontainers.GenericContainer(context.Background(), req)
	testcontainers.CleanupContainer(t, ctr1)
	require.NoError(t, err)
	require.NotNil(t, ctr1)
}

func TestGenericContainer_pause_start_withReuse(t *testing.T) {
	containerName := "my-nginx"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        nginxAlpineImage,
			ExposedPorts: []string{"8080/tcp"},
			Name:         containerName,
		},
		Reuse:   true,
		Started: true,
	}

	ctr, err := testcontainers.GenericContainer(context.Background(), req)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.NotNil(t, ctr)

	// Pause the container is not supported by our API, but we can do it manually
	// by using the Docker client.
	cli, err := core.NewClient(context.Background())
	require.NoError(t, err)

	err = cli.ContainerPause(context.Background(), ctr.GetContainerID())
	require.NoError(t, err)

	// Because the container is paused, it should not be possible to start it again.
	ctr1, err := testcontainers.GenericContainer(context.Background(), req)
	testcontainers.CleanupContainer(t, ctr1)
	require.ErrorIs(t, err, errors.ErrUnsupported)
}
