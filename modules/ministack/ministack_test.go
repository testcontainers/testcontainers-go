package ministack

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	ministackImage = "ministackorg/ministack:1.5.45"
)

func generateContainerRequest() *testcontainers.GenericContainerRequest {
	return &testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env:          map[string]string{},
			ExposedPorts: []string{},
		},
	}
}

func TestSetDockerHost(t *testing.T) {
	tests := []struct {
		envVar string
	}{
		{hostnameExternalEnvVar},
		{ministackHostEnvVar},
	}

	for _, tt := range tests {
		t.Run("HOSTNAME_EXTERNAL variable is passed as part of the request", func(t *testing.T) {
			req := generateContainerRequest()

			req.Env[tt.envVar] = "foo"

			reason, err := setDockerHost(context.Background(), req, tt.envVar)
			require.NoError(t, err)
			require.Equal(t, "explicitly as environment variable", reason)
			require.Equal(t, "foo", req.Env[tt.envVar])
		})

		t.Run("HOSTNAME_EXTERNAL matches the last network alias on a container with non-default network", func(t *testing.T) {
			req := generateContainerRequest()

			req.Networks = []string{"foo", "bar", "baaz"}
			req.NetworkAliases = map[string][]string{
				"foo":  {"foo0", "foo1", "foo2", "foo3"},
				"bar":  {"bar0", "bar1", "bar2", "bar3"},
				"baaz": {"baaz0", "baaz1", "baaz2", "baaz3"},
			}

			reason, err := setDockerHost(context.Background(), req, tt.envVar)
			require.NoError(t, err)
			require.Equal(t, "to match last network alias on container with non-default network", reason)
			require.Equal(t, "foo3", req.Env[tt.envVar])
		})

		t.Run("HOSTNAME_EXTERNAL matches the daemon host because there are no aliases", func(t *testing.T) {
			dockerProvider, err := testcontainers.NewDockerProvider()
			require.NoError(t, err)
			defer dockerProvider.Close()

			// because the daemon host could be a remote one, we need to get it from the provider
			expectedDaemonHost, err := dockerProvider.DaemonHost(context.Background())
			require.NoError(t, err)

			req := generateContainerRequest()

			req.Networks = []string{"foo", "bar", "baaz"}
			req.NetworkAliases = map[string][]string{}

			reason, err := setDockerHost(context.Background(), req, tt.envVar)
			require.NoError(t, err)
			require.Equal(t, "to match host-routable address for container", reason)
			require.Equal(t, expectedDaemonHost, req.Env[tt.envVar])
		})
	}
}

func TestRunContainer(t *testing.T) {
	tests := []struct {
		version string
	}{
		{"1.5.45"},
		{"latest"},
		{"full"},
	}

	for _, tt := range tests {
		ctx := context.Background()

		ctr, err := Run(
			ctx,
			"ministackorg/ministack:"+tt.version,
		)
		testcontainers.CleanupContainer(t, ctr)

		t.Run("ministack:"+tt.version+" - multiple services exposed on same port", func(t *testing.T) {
			require.NoError(t, err)
			require.NotNil(t, ctr)

			inspect, err := ctr.Inspect(ctx)
			require.NoError(t, err)

			rawPorts := inspect.NetworkSettings.Ports

			ports := 0
			// only one port is exposed among all the ports in the container
			for _, v := range rawPorts {
				if len(v) > 0 {
					ports++
				}
			}

			require.Equal(t, 1, ports) // a single port is exposed
		})
	}
}

func TestStartWithoutOverride(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, ministackImage)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.NotNil(t, ctr)
}

func TestStartV2WithNetwork(t *testing.T) {
	ctx := context.Background()

	nw, err := network.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, nw)

	ministack, err := Run(
		ctx,
		ministackImage,
		network.WithNetwork([]string{"ministack"}, nw),
		testcontainers.WithEnv(map[string]string{"SERVICES": "s3,sqs"}),
	)
	testcontainers.CleanupContainer(t, ministack)
	require.NoError(t, err)
	require.NotNil(t, ministack)

	cli, err := testcontainers.Run(
		ctx, "amazon/aws-cli:2.42.5",
		network.WithNetwork([]string{"cli"}, nw),
		testcontainers.WithEntrypoint("tail"),
		testcontainers.WithCmd("-f", "/dev/null"),
		testcontainers.WithEnv(map[string]string{
			"AWS_ACCESS_KEY_ID":     "accesskey",
			"AWS_SECRET_ACCESS_KEY": "secretkey",
			"AWS_REGION":            "eu-west-1",
		}),
		testcontainers.WithWaitStrategy(wait.ForExec([]string{
			"/usr/local/bin/aws", "sqs", "create-queue", "--queue-name", "baz", "--region", "eu-west-1",
			"--endpoint-url", "http://ministack:4566", "--no-verify-ssl",
		}).
			WithStartupTimeout(time.Second*10).
			WithExitCodeMatcher(func(exitCode int) bool {
				return exitCode == 0
			}).
			WithResponseMatcher(func(r io.Reader) bool {
				respBytes, _ := io.ReadAll(r)
				resp := string(respBytes)
				return strings.Contains(resp, "http://ministack:4566")
			}),
		),
	)
	testcontainers.CleanupContainer(t, cli)
	require.NoError(t, err)
	require.NotNil(t, cli)
}
