package testcontainers_test

import (
	"context"
	"io"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestOverrideContainerRequest(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"BAR": "BAR",
			},
			Image:        "foo",
			ExposedPorts: []string{"12345/tcp"},
			WaitingFor: wait.ForNop(
				func(ctx context.Context, target wait.StrategyTarget) error {
					return nil
				},
			),
			Networks: []string{"foo", "bar", "baaz"},
			NetworkAliases: map[string][]string{
				"foo": {"foo0", "foo1", "foo2", "foo3"},
			},
		},
	}

	toBeMergedRequest := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"FOO": "FOO",
			},
			Image:        "bar",
			ExposedPorts: []string{"67890/tcp"},
			Networks:     []string{"foo1", "bar1"},
			NetworkAliases: map[string][]string{
				"foo1": {"bar"},
			},
			WaitingFor: wait.ForLog("foo"),
		},
	}

	// the toBeMergedRequest should be merged into the req
	testcontainers.CustomizeRequest(toBeMergedRequest)(&req)

	// toBeMergedRequest should not be changed
	assert.Equal(t, "", toBeMergedRequest.Env["BAR"])
	assert.Equal(t, 1, len(toBeMergedRequest.ExposedPorts))
	assert.Equal(t, "67890/tcp", toBeMergedRequest.ExposedPorts[0])

	// req should be merged with toBeMergedRequest
	assert.Equal(t, "FOO", req.Env["FOO"])
	assert.Equal(t, "BAR", req.Env["BAR"])
	assert.Equal(t, "bar", req.Image)
	assert.Equal(t, []string{"12345/tcp", "67890/tcp"}, req.ExposedPorts)
	assert.Equal(t, []string{"foo", "bar", "baaz", "foo1", "bar1"}, req.Networks)
	assert.Equal(t, []string{"foo0", "foo1", "foo2", "foo3"}, req.NetworkAliases["foo"])
	assert.Equal(t, []string{"bar"}, req.NetworkAliases["foo1"])
	assert.Equal(t, wait.ForLog("foo"), req.WaitingFor)
}

func TestWithNetwork(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{},
	}

	testcontainers.WithNetwork("new-network", "alias")(&req)

	assert.Equal(t, []string{"new-network"}, req.Networks)
	assert.Equal(t, map[string][]string{"new-network": {"alias"}}, req.NetworkAliases)

	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	args := filters.NewArgs()
	args.Add("name", "new-network")

	resources, err := client.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: args,
	})
	require.NoError(t, err)
	assert.Len(t, resources, 1)

	assert.Equal(t, "new-network", resources[0].Name)
}

func TestWithNetworkMultipleCallsWithSameNameReuseTheNetwork(t *testing.T) {
	for int := 0; int < 100; int++ {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{},
		}

		testcontainers.WithNetwork("new-network", "alias")(&req)
		assert.Equal(t, []string{"new-network"}, req.Networks)
		assert.Equal(t, map[string][]string{"new-network": {"alias"}}, req.NetworkAliases)
	}

	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	args := filters.NewArgs()
	args.Add("name", "new-network")

	resources, err := client.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: args,
	})
	require.NoError(t, err)
	assert.Len(t, resources, 1)

	assert.Equal(t, "new-network", resources[0].Name)
}

func TestWithStartupCommand(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      "alpine",
			Entrypoint: []string{"tail", "-f", "/dev/null"},
		},
		Started: true,
	}

	testExec := testcontainers.NewRawCommand([]string{"touch", "/tmp/.testcontainers"})

	testcontainers.WithStartupCommand(testExec)(&req)

	assert.Equal(t, 1, len(req.LifecycleHooks))
	assert.Equal(t, 1, len(req.LifecycleHooks[0].PostStarts))

	c, err := testcontainers.GenericContainer(context.Background(), req)
	require.NoError(t, err)
	defer func() {
		err = c.Terminate(context.Background())
		require.NoError(t, err)
	}()

	_, reader, err := c.Exec(context.Background(), []string{"ls", "/tmp/.testcontainers"}, exec.Multiplexed())
	require.NoError(t, err)

	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/.testcontainers\n", string(content))
}
