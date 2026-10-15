package testcontainers

import (
	"context"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

// A child leaves a stopped container, network and named volume behind. Only
// Ryuk can make this test pass: explicit cleanup happens after a failed check.
func TestReaperReapsOrphans(t *testing.T) {
	if config.Read().RyukDisabled {
		t.Skip("Ryuk is disabled")
	}
	if os.Getenv("TCGO_REAPER_CHILD") == "1" {
		createReaperOrphans(t)
		return
	}

	ctx := t.Context()
	cli, err := NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cli.Close() })
	session := uuid.NewString()
	filters := make(client.Filters).Add("label", core.LabelSessionID+"="+session)
	t.Cleanup(func() {
		// Collect evidence before fallback cleanup, including Ryuk's logs.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		containers, listErr := cli.ContainerList(cleanupCtx, client.ContainerListOptions{All: true, Filters: filters})
		if listErr != nil {
			t.Errorf("list orphan containers for cleanup: %v", listErr)
		}
		for _, ctr := range containers.Items {
			t.Logf("orphan container: %s, state: %s, labels: %v", ctr.ID, ctr.State, ctr.Labels)
			logs, logErr := cli.ContainerLogs(cleanupCtx, ctr.ID, client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Tail: "100"})
			if logErr == nil {
				body, _ := io.ReadAll(logs)
				_ = logs.Close()
				t.Logf("container logs: %s", body)
			}
			_, removeErr := cli.ContainerRemove(cleanupCtx, ctr.ID, client.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
			require.NoError(t, removeErr)
		}
		networks, listErr := cli.NetworkList(cleanupCtx, client.NetworkListOptions{Filters: filters})
		require.NoError(t, listErr)
		for _, network := range networks.Items {
			t.Logf("orphan network: %s", network.Name)
			_, removeErr := cli.NetworkRemove(cleanupCtx, network.ID, client.NetworkRemoveOptions{})
			require.NoError(t, removeErr)
		}
		volumes, listErr := cli.VolumeList(cleanupCtx, client.VolumeListOptions{Filters: filters})
		require.NoError(t, listErr)
		for _, volume := range volumes.Items {
			t.Logf("orphan volume: %s", volume.Name)
			_, removeErr := cli.VolumeRemove(cleanupCtx, volume.Name, client.VolumeRemoveOptions{Force: true})
			require.NoError(t, removeErr)
		}
	})

	binary, err := os.Executable()
	require.NoError(t, err)
	childCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(childCtx, binary, "-test.run=^TestReaperReapsOrphans$", "-test.v")
	cmd.Env = append(os.Environ(), "TCGO_REAPER_CHILD=1", "TESTCONTAINERS_SESSION_ID="+session)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "orphan child: %s", output)

	// The child has exited, closing all of its Ryuk keepalive connections.
	pollCtx, pollCancel := context.WithTimeout(ctx, time.Minute)
	defer pollCancel()
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()
	for {
		containers, err := cli.ContainerList(pollCtx, client.ContainerListOptions{All: true, Filters: filters})
		require.NoError(t, err)
		networks, err := cli.NetworkList(pollCtx, client.NetworkListOptions{Filters: filters})
		require.NoError(t, err)
		volumes, err := cli.VolumeList(pollCtx, client.VolumeListOptions{Filters: filters})
		require.NoError(t, err)
		if len(containers.Items)+len(networks.Items)+len(volumes.Items) == 0 {
			return
		}
		select {
		case <-pollCtx.Done():
			t.Fatalf("Ryuk left %d containers, %d networks and %d volumes for session %s; child: %s", len(containers.Items), len(networks.Items), len(volumes.Items), session, output)
		case <-tick.C:
		}
	}
}

func createReaperOrphans(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	selection := ProviderDefault
	if os.Getenv("TCGO_PODMAN_COMPAT") == "explicit" {
		selection = ProviderPodman
	}
	p, err := selection.GetProvider()
	require.NoError(t, err)
	defer p.Close()
	session := config.Read().SessionID
	cli, err := NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer cli.Close()
	// Podman's compatibility API ignores Mount.VolumeOptions.Labels. Create a
	// labelled volume explicitly so this checks Ryuk's volume removal as well.
	_, err = cli.VolumeCreate(ctx, client.VolumeCreateOptions{
		Name: "orphan-" + session, Labels: GenericLabels(),
	})
	require.NoError(t, err)
	// Start Ryuk through ordinary library creation. Do not register any explicit
	// resource cleanup: the parent must observe deletion after this child exits.
	ctr, err := p.RunContainer(ctx, ContainerRequest{
		Image:  "alpine:3.23",
		Cmd:    []string{"sleep", "300"},
		Mounts: Mounts(VolumeMount("orphan-"+session, "/data")),
	})
	require.NoError(t, err)
	stopTimeout := time.Second
	require.NoError(t, ctr.Stop(ctx, &stopTimeout))
	state, err := ctr.State(ctx)
	require.NoError(t, err)
	require.False(t, state.Running)
	_, err = p.CreateNetwork(ctx, NetworkRequest{Name: "orphan-" + session})
	require.NoError(t, err)
	filters := make(client.Filters).Add("label", core.LabelSessionID+"="+session).
		Add("label", core.LabelReap+"=true")
	containers, err := cli.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: filters})
	require.NoError(t, err)
	require.Len(t, containers.Items, 1)
	networks, err := cli.NetworkList(ctx, client.NetworkListOptions{Filters: filters})
	require.NoError(t, err)
	require.Len(t, networks.Items, 1)
	volumes, err := cli.VolumeList(ctx, client.VolumeListOptions{Filters: filters})
	require.NoError(t, err)
	require.Len(t, volumes.Items, 1)
}
