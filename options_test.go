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
)

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

type testExecutable struct {
	cmds []string
}

func (t testExecutable) AsCommand() []string {
	return t.cmds
}

func TestWithStartupCommand(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      "alpine",
			Entrypoint: []string{"tail", "-f", "/dev/null"},
		},
		Started: true,
	}

	testExec := testExecutable{
		cmds: []string{"touch", "/tmp/.testcontainers"},
	}

	testcontainers.WithStartupCommand(testExec)(&req)

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
