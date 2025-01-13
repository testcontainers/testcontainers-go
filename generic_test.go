package testcontainers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	reusableMarkerTestLabel string = "TEST_REUSABLE_MARKER"
)

var reusableReq = ContainerRequest{
	Image:        nginxDelayedImage,
	ExposedPorts: []string{nginxDefaultPort},
	WaitingFor:   wait.ForListeningPort(nginxDefaultPort), // default startupTimeout is 60s
	Reuse:        true,
}

func TestGenericReusableContainer_reused(t *testing.T) {
	ctx := context.Background()

	n1, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: reusableReq,
		Started:          true,
	})
	require.NoError(t, err)
	require.True(t, n1.IsRunning())
	CleanupContainer(t, n1)

	copiedFileName := "hello_copy.sh"
	err = n1.CopyFileToContainer(ctx, "./testdata/hello.sh", "/"+copiedFileName, 700)
	require.NoError(t, err)

	n2, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: reusableReq,
		Started:          true,
	})
	require.NoError(t, err)

	c, _, err := n2.Exec(ctx, []string{"/bin/sh", copiedFileName})
	require.NoError(t, err)
	require.Zero(t, c)
}

func TestGenericReusableContainer_notReused(t *testing.T) {
	ctx := context.Background()

	n1, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: reusableReq,
		Started:          true,
	})
	require.NoError(t, err)
	require.True(t, n1.IsRunning())
	CleanupContainer(t, n1)

	copiedFileName := "hello_copy.sh"
	err = n1.CopyFileToContainer(ctx, "./testdata/hello.sh", "/"+copiedFileName, 700)
	require.NoError(t, err)

	// because the second container is not marked for reuse, a new container will be created
	// even though the hashes are the same.
	old := reusableReq
	t.Cleanup(func() {
		reusableReq = old
	})
	reusableReq.Hostname = "foo"

	n2, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: reusableReq,
		Started:          true,
	})
	require.NoError(t, err)
	CleanupContainer(t, n2)

	c, _, err := n2.Exec(ctx, []string{"/bin/sh", copiedFileName})
	require.NoError(t, err)
	require.NotZero(t, c) // the file does not exist in the new container, so it must fail
}

func TestGenericContainerShouldReturnRefOnError(t *testing.T) {
	// In this test, we are going to cancel the context to exit the `wait.Strategy`.
	// We want to make sure that the GenericContainer call will still return a reference to the
	// created container, so that we can Destroy it.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:      nginxAlpineImage,
			WaitingFor: wait.ForLog("this string should not be present in the logs"),
		},
		Started: true,
	})
	require.Error(t, err)
	require.NotNil(t, c)
	CleanupContainer(t, c)
}

func TestGenericReusableContainerInSubprocess(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(10)

	creatingMessage := "🐳 Creating container for image " + nginxDelayedImage
	creatingCount := 0

	reusingMessage := "🔥 Container reused"
	reusingCount := 0

	minCreatedCount := 1
	maxReusedCount := 9
	totalCount := minCreatedCount + maxReusedCount

	for i := 0; i < totalCount; i++ {
		go func() {
			defer wg.Done()

			// create containers in subprocesses, as "go test ./..." does.
			output := createReuseContainerInSubprocess(t)

			t.Log(output)

			if strings.Contains(output, creatingMessage) {
				creatingCount++
			}

			if strings.Contains(output, reusingMessage) {
				reusingCount++
			}
		}()
	}

	wg.Wait()

	// because we cannot guarantee the daemon will reuse the container, we can only assert that
	// the container was created at least once and reused at least once. This flakiness is due to
	// the fact that the code is checking for a few seconds if the container with the hash labels is
	// already running, and because this test is using separate test processes, it could be possible
	// that the container is not reused in time.
	// This flakiness is documented in the Reuse docs, so this test demonstrates that is usually working.
	t.Logf("Created: %v, Reused: %v, Total: %v", creatingCount, reusingCount, totalCount)

	require.LessOrEqual(t, creatingCount, totalCount)
	require.LessOrEqual(t, reusingCount, totalCount)
	require.Equal(t, totalCount, reusingCount+creatingCount)

	// cleanup the containers that could have been created in the subprocesses

	cli, err := core.NewClient(context.Background())
	require.NoError(t, err)

	// We need to find the containers created in the subprocesses and terminate them.
	// For that, we are going to search for containers with the reusable test label.
	f := filters.NewArgs(filters.Arg("label", fmt.Sprintf("%s=%s", reusableMarkerTestLabel, "true")))
	cs, err := cli.ContainerList(context.Background(), container.ListOptions{Filters: f})
	require.NoError(t, err)
	defer cli.Close()

	provider, err := NewDockerProvider()
	require.NoError(t, err)

	provider.SetClient(cli)

	for _, c := range cs {
		nginxC, err := provider.ContainerFromType(context.Background(), c)
		CleanupContainer(t, nginxC)
		require.NoError(t, err)
	}
}

func createReuseContainerInSubprocess(t *testing.T) string {
	t.Helper()
	// force verbosity in subprocesses, so that the output is printed
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperContainerStarterProcess", "-test.v=true")
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

	// we are going to mark the container with a test label, so that we can find it later
	req := reusableReq
	if req.Labels == nil {
		req.Labels = map[string]string{}
	}
	req.Labels[reusableMarkerTestLabel] = "true"

	nginxC, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	require.True(t, nginxC.IsRunning())
}
