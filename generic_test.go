package testcontainers

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	reusableContainerName = "my_test_reusable_container"
)

func TestGenericReusableContainer(t *testing.T) {
	ctx := context.Background()

	reusableContainerName := reusableContainerName + "_" + time.Now().Format("20060102150405")

	n1, err := Run(ctx, nginxAlpineImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
		WithName(reusableContainerName),
	)
	require.NoError(t, err)
	require.True(t, n1.IsRunning())
	CleanupContainer(t, n1)

	copiedFileName := "hello_copy.sh"
	err = n1.CopyFileToContainer(ctx, "./testdata/hello.sh", "/"+copiedFileName, 700)
	require.NoError(t, err)

	tests := []struct {
		name          string
		containerName string
		errorMatcher  func(t *testing.T, err error)
		reuseOption   bool
	}{
		{
			name: "reuse option with empty name",
			errorMatcher: func(t *testing.T, err error) {
				t.Helper()
				require.ErrorContains(t, err, "container name must be provided")
			},
			reuseOption: true,
		},
		{
			name:          "container already exists (reuse=false)",
			containerName: reusableContainerName,
			errorMatcher: func(t *testing.T, err error) {
				t.Helper()
				require.Error(t, err)
			},
			reuseOption: false,
		},
		{
			name:          "success reusing",
			containerName: reusableContainerName,
			reuseOption:   true,
			errorMatcher: func(t *testing.T, err error) {
				t.Helper()
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := []ContainerCustomizer{
				WithExposedPorts(nginxDefaultPort),
				WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)),
				WithName(tc.containerName),
			}
			if tc.reuseOption {
				opts = append(opts, WithReuseByName(tc.containerName))
			}

			n2, err := Run(ctx, nginxAlpineImage, opts...)
			CleanupContainer(t, n2)
			tc.errorMatcher(t, err)

			if err == nil {
				c, _, err := n2.Exec(ctx, []string{"/bin/ash", copiedFileName})
				require.NoError(t, err)
				require.Zero(t, c)
			}
		})
	}
}

func TestGenericContainerShouldReturnRefOnError(t *testing.T) {
	// In this test, we are going to cancel the context to exit the `wait.Strategy`.
	// We want to make sure that the Run call will still return a reference to the
	// created container, so that we can Destroy it.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c, err := Run(ctx, nginxAlpineImage, WithWaitStrategy(wait.ForLog("this string should not be present in the logs")))
	CleanupContainer(t, c)
	require.Error(t, err)
	require.NotNil(t, c)
}

func TestGenericReusableContainerInSubprocess(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(10)
	for range 10 {
		go func() {
			defer wg.Done()

			// create containers in subprocesses, as "go test ./..." does.
			output := createReuseContainerInSubprocess(t)

			t.Log(output)
			// check is reuse container with WaitingFor work correctly.
			require.Contains(t, output, "⏳ Waiting for container id")
			require.Contains(t, output, "🔔 Container is ready")
		}()
	}

	wg.Wait()

	cli, err := NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	f := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: reusableContainerName})

	ctrs, err := cli.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: f,
	})
	require.NoError(t, err)
	require.Len(t, ctrs, 1)

	provider, err := NewDockerProvider()
	require.NoError(t, err)

	provider.SetClient(cli)

	nginxC, err := provider.ContainerFromType(context.Background(), ctrs[0])
	CleanupContainer(t, nginxC)
	require.NoError(t, err)
}

func createReuseContainerInSubprocess(t *testing.T) string {
	t.Helper()
	// force verbosity in subprocesses, so that the output is printed
	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=TestHelperContainerStarterProcess", "-test.v=true")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	return string(output)
}

// TestHelperContainerStarterProcess is a helper function
// to start a container in a subprocess. It's not a real test.
func TestHelperContainerStarterProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		t.Skip("Skipping helper test function. It's not a real test")
	}

	ctx := context.Background()

	nginxC, err := Run(ctx, nginxDelayedImage,
		WithExposedPorts(nginxDefaultPort),
		WithWaitStrategy(wait.ForListeningPort(nginxDefaultPort)), // default startupTimeout is 60s
		WithReuseByName(reusableContainerName),
	)
	require.NoError(t, err)
	require.True(t, nginxC.IsRunning())
}
