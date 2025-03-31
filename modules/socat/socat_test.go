package socat_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/socat"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestSocat(t *testing.T) {
	ctx := context.Background()

	ctr, err := socat.Run(ctx, "alpine/socat:1.8.0.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}

func TestRun_helloWorld(t *testing.T) {
	ctx := context.Background()

	nw, err := network.New(ctx)
	testcontainers.CleanupNetwork(t, nw)
	require.NoError(t, err)

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "testcontainers/helloworld:1.2.0",
			ExposedPorts: []string{"8080/tcp"},
			Networks:     []string{nw.Name},
			NetworkAliases: map[string][]string{
				nw.Name: {"helloworld"},
			},
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	const exposedPort = 8080

	target := socat.NewTarget(exposedPort, "helloworld")

	socatContainer, err := socat.Run(
		ctx, "alpine/socat:1.8.0.1",
		socat.WithTargets(target),
		network.WithNetwork([]string{"socat"}, nw),
	)
	testcontainers.CleanupContainer(t, socatContainer)
	require.NoError(t, err)

	httpClient := http.DefaultClient

	baseURI := socatContainer.TargetURL(exposedPort)
	require.NotNil(t, baseURI)

	resp, err := httpClient.Get(baseURI.String() + "/ping")
	require.NoError(t, err)

	require.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "PONG", string(body))
}

func TestRun_helloWorldDifferentPort(t *testing.T) {
	ctx := context.Background()

	nw, err := network.New(ctx)
	testcontainers.CleanupNetwork(t, nw)
	require.NoError(t, err)

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "testcontainers/helloworld:1.2.0",
			ExposedPorts: []string{"8080/tcp"},
			Networks:     []string{nw.Name},
			NetworkAliases: map[string][]string{
				nw.Name: {"helloworld"},
			},
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	const (
		// The helloworld container is listening on both ports: 8080 and 8081
		exposedPort  = 8080
		internalPort = 8081
	)

	target := socat.NewTargetWithInternalPort(exposedPort, internalPort, "helloworld")

	socatContainer, err := socat.Run(
		ctx, "alpine/socat:1.8.0.1",
		socat.WithTargets(target),
		network.WithNetwork([]string{"socat"}, nw),
	)
	testcontainers.CleanupContainer(t, socatContainer)
	require.NoError(t, err)

	httpClient := http.DefaultClient

	baseURI := socatContainer.TargetURL(exposedPort)
	require.NotNil(t, baseURI)

	resp, err := httpClient.Get(baseURI.String() + "/ping")
	require.NoError(t, err)

	require.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "PONG", string(body))
}
