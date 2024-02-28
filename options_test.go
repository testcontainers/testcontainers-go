package testcontainers_test

import (
	"context"
	"io"
	"testing"

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
	assert.Len(t, toBeMergedRequest.ExposedPorts, 1)
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

type msgsLogConsumer struct {
	msgs []string
}

// Accept prints the log to stdout
func (lc *msgsLogConsumer) Accept(l testcontainers.Log) {
	lc.msgs = append(lc.msgs, string(l.Content))
}

func TestWithLogConsumers(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      "mysql:8.0.36",
			WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL"),
		},
		Started: true,
	}

	lc := &msgsLogConsumer{}

	testcontainers.WithLogConsumers(lc)(&req)

	c, err := testcontainers.GenericContainer(context.Background(), req)
	// we expect an error because the MySQL environment variables are not set
	// but this is expected because we just want to test the log consumer
	require.Error(t, err)
	defer func() {
		err = c.Terminate(context.Background())
		require.NoError(t, err)
	}()

	assert.NotEmpty(t, lc.msgs)
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

	assert.Len(t, req.LifecycleHooks, 1)
	assert.Len(t, req.LifecycleHooks[0].PostStarts, 1)

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

func TestWithAfterReadyCommand(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      "alpine",
			Entrypoint: []string{"tail", "-f", "/dev/null"},
		},
		Started: true,
	}

	testExec := testcontainers.NewRawCommand([]string{"touch", "/tmp/.testcontainers"})

	testcontainers.WithAfterReadyCommand(testExec)(&req)

	assert.Len(t, req.LifecycleHooks, 1)
	assert.Len(t, req.LifecycleHooks[0].PostReadies, 1)

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

func TestWithEnv(t *testing.T) {
	tests := map[string]struct {
		req    *testcontainers.GenericContainerRequest
		env    map[string]string
		expect map[string]string
	}{
		"add": {
			req: &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Env: map[string]string{"KEY1": "VAL1"},
				},
			},
			env: map[string]string{"KEY2": "VAL2"},
			expect: map[string]string{
				"KEY1": "VAL1",
				"KEY2": "VAL2",
			},
		},
		"add-nil": {
			req:    &testcontainers.GenericContainerRequest{},
			env:    map[string]string{"KEY2": "VAL2"},
			expect: map[string]string{"KEY2": "VAL2"},
		},
		"override": {
			req: &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Env: map[string]string{
						"KEY1": "VAL1",
						"KEY2": "VAL2",
					},
				},
			},
			env: map[string]string{"KEY2": "VAL3"},
			expect: map[string]string{
				"KEY1": "VAL1",
				"KEY2": "VAL3",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			opt := testcontainers.WithEnv(tc.env)
			opt.Customize(tc.req)
			require.Equal(t, tc.expect, tc.req.Env)
		})
	}
}
