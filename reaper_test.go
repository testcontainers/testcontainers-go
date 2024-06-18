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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/internal/core/reaper"
)

const (
	// testSessionID the tests need to create a reaper in a different session, so that it does not interfere with other tests
	testSessionID string = "this-is-a-different-session-id"

	// testSessionFromTestProgram the tests need to create a reaper in a different session, so that it does not interfere with other tests
	testSessionFromTestProgram string = "reusing-reaper-from-other-test-program-using-docker"
)

func TestContainerStartsWithoutTheReaper(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	if !config.Read().RyukDisabled {
		t.Skip("Ryuk is enabled, skipping test")
	}

	ctx := context.Background()

	ctr, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		Started: true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, ctr)

	sessionID := core.SessionID()

	reaperContainer, err := lookUpReaperContainer(ctx, sessionID)
	if err != nil {
		t.Fatal(err, "expected reaper container not found.")
	}
	if reaperContainer != nil {
		t.Fatal("expected zero reaper running.")
	}
}

func TestContainerStartsWithTheReaper(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	if config.Read().RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	ctx := context.Background()

	c, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	TerminateContainerOnEnd(t, ctx, c)

	sessionID := core.SessionID()

	reaperContainer, err := lookUpReaperContainer(ctx, sessionID)
	if err != nil {
		t.Fatal(err, "expected reaper container running.")
	}
	if reaperContainer == nil {
		t.Fatal("expected one reaper to be running.")
	}
}

func TestContainerStopWithReaper(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	if config.Read().RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	ctx := context.Background()

	nginxA, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		Started: true,
	})

	require.NoError(t, err)
	TerminateContainerOnEnd(t, ctx, nginxA)

	state, err := nginxA.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Running != true {
		t.Fatal("The container shoud be in running state")
	}
	stopTimeout := 10 * time.Second
	err = nginxA.Stop(ctx, &stopTimeout)
	if err != nil {
		t.Fatal(err)
	}

	state, err = nginxA.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Running != false {
		t.Fatal("The container shoud not be running")
	}
	if state.Status != "exited" {
		t.Fatal("The container shoud be in exited state")
	}
}

func TestContainerTerminationWithReaper(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	if config.Read().RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	ctx := context.Background()

	nginxA, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	state, err := nginxA.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Running != true {
		t.Fatal("The container shoud be in running state")
	}
	err = nginxA.Terminate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = nginxA.State(ctx)
	if err == nil {
		t.Fatal("expected error from container inspect.")
	}
}

func TestContainerTerminationWithoutReaper(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	if config.Read().RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	ctx := context.Background()

	nginxA, err := New(ctx, Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	state, err := nginxA.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Running != true {
		t.Fatal("The container shoud be in running state")
	}
	err = nginxA.Terminate(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = nginxA.State(ctx)
	if err == nil {
		t.Fatal("expected error from container inspect.")
	}
}

func TestReaperReusedIfHealthy(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	if config.Read().RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	SkipIfContainerRuntimeIsNotHealthy(&testing.T{})

	ctx := context.Background()

	// because Ryuk is not disabled, the returned reaper is not nil
	r, err := NewReaper(ctx, testSessionID)
	require.NoError(t, err, "creating the Reaper should not error")

	reaperReused, err := NewReaper(ctx, testSessionID)
	require.NoError(t, err, "reusing the Reaper should not error")
	// assert that the internal state of both reaper instances is the same
	assert.Equal(t, r.SessionID, reaperReused.SessionID, "expecting the same SessionID")
	assert.Equal(t, r.Endpoint, reaperReused.Endpoint, "expecting the same reaper endpoint")
	assert.Equal(t, r.Container.GetContainerID(), reaperReused.Container.GetContainerID(), "expecting the same container ID")
	assert.Equal(t, r.SessionID, reaperReused.SessionID, "expecting the same session ID")

	terminate, err := reaper.Connect()
	defer func(term chan bool) {
		term <- true
	}(terminate)
	require.NoError(t, err, "connecting to Reaper should be successful")
}

func TestRecreateReaperIfTerminated(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	tcConfig := config.Read()
	if tcConfig.RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	SkipIfContainerRuntimeIsNotHealthy(&testing.T{})

	ctx := context.Background()

	// because Ryuk is not disabled, the returned reaper is not nil
	r, err := NewReaper(ctx, testSessionID)
	require.NoError(t, err, "creating the Reaper should not error")

	// Wait for ryuk's default timeout (10s) + 1s to allow for a graceful shutdown/cleanup of the container.
	time.Sleep(11 * time.Second)

	reaperInstance = nil
	reaperOnce = sync.Once{}

	recreatedReaper, err := NewReaper(ctx, testSessionID)
	require.NoError(t, err, "creating the Reaper should not error")
	assert.NotEqual(t, r.Container.GetContainerID(), recreatedReaper.Container.GetContainerID(), "expected different container ID")

	terminate, err := reaper.Connect()
	defer func(term chan bool) {
		term <- true
	}(terminate)
	require.NoError(t, err, "connecting to Reaper should be successful")
}

func TestReaper_reuseItFromOtherTestProgramUsingDocker(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	if config.Read().RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	occurrences := 5
	// create different test calls in subprocesses, with the same session ID, as "go test ./..." does.
	// The test will simply call NewReaper, which should return the same reaper container instance.
	output := createReaperContainerInSubprocess(t, occurrences)

	// check if reaper container is obtained from Docker the number of occurrences minus one times: the first time it's created
	assert.Equal(t, occurrences-1, strings.Count(output, "🔥 Reaper obtained from this test session"), output)
}

func createReaperContainerInSubprocess(t *testing.T, occurrences int) string {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperReaperFromOtherProgram", fmt.Sprintf("-test.count=%d", occurrences))
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	return string(output)
}

// TestHelperContainerStarterProcess is a helper function
// to start a container in a subprocess. It's not a real test.
func TestHelperReaperFromOtherProgram(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		t.Skip("Skipping helper test function. It's not a real test")
	}

	ctx := context.Background()

	_, err := NewReaper(ctx, testSessionFromTestProgram)
	require.NoError(t, err, "creating the Reaper should not error")
}

// TestReaper_ReuseRunning tests whether reusing the reaper if using
// testcontainers from concurrently multiple packages works as expected. In this
// case, global locks are without any effect as Go tests different packages
// isolated. Therefore, this test does not use the same logic with locks on
// purpose. We expect reaper creation to still succeed in case a reaper is
// already running for the same session id by returning its container instance
// instead.
func TestReaper_ReuseRunning(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	tcConfig := config.Read()
	if tcConfig.RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	const concurrency = 64

	timeout, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sessionID := core.SessionID()

	obtainedReaperContainerIDs := make([]string, concurrency)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			// because Ryuk is not disabled, the returned reaper is not nil
			createdReaper, err := NewReaper(timeout, sessionID)
			require.NoError(t, err, "new reaper should not fail")
			obtainedReaperContainerIDs[i] = createdReaper.Container.GetContainerID()
		}()
	}
	wg.Wait()

	// Assure that all calls returned the same container.
	firstContainerID := obtainedReaperContainerIDs[0]
	for i, containerID := range obtainedReaperContainerIDs {
		assert.Equal(t, firstContainerID, containerID, "call %d should have returned same container id", i)
	}
}
