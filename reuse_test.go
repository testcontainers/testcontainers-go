package testcontainers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestGenericContainer_stop_start_withReuse(t *testing.T) {
	containerName := "my-postgres"

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
