package testcontainers

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testSessionID the tests need to create a reaper in a different session, so that it does not interfere with other tests
const testSessionID = "this-is-a-different-session-id"

type mockReaperProvider struct {
	req              ContainerRequest
	hostConfig       *container.HostConfig
	endpointSettings map[string]*network.EndpointSettings
	config           TestcontainersConfig
}

func newMockReaperProvider(cfg config.Config) *mockReaperProvider {
	m := &mockReaperProvider{
		config: TestcontainersConfig{
			Config: cfg,
		},
	}

	return m
}

var errExpected = errors.New("expected")

func (m *mockReaperProvider) RunContainer(_ context.Context, req ContainerRequest) (Container, error) {
	m.req = req

	m.hostConfig = &container.HostConfig{}
	m.endpointSettings = map[string]*network.EndpointSettings{}

	if req.HostConfigModifier == nil {
		req.HostConfigModifier = defaultHostConfigModifier(req)
	}
	req.HostConfigModifier(m.hostConfig)

	if req.EndpointSettingsModifier != nil {
		req.EndpointSettingsModifier(m.endpointSettings)
	}

	// we're only interested in the request, so instead of mocking the Docker client
	// we'll error out here
	return nil, errExpected
}

func (m *mockReaperProvider) Config() TestcontainersConfig {
	return m.config
}

// expectedReaperRequest creates the expected reaper container request with the given customizations.
func expectedReaperRequest(customize ...func(*ContainerRequest)) ContainerRequest {
	req := ContainerRequest{
		Image:        config.ReaperDefaultImage,
		ExposedPorts: []string{"8080/tcp"},
		Labels:       core.DefaultLabels(testSessionID),
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.Binds = []string{core.MustExtractDockerSocket(context.Background()) + ":/var/run/docker.sock"}
			hostConfig.Privileged = true
		},
		WaitingFor: newReaperSpawner().defaultWaitStrategy(testSessionID, &reaperConn{}),
		Env: map[string]string{
			"RYUK_PORT":                 "8080",
			"RYUK_CONNECTION_TIMEOUT":   "1m0s",
			"RYUK_RECONNECTION_TIMEOUT": "10s",
		},
	}

	req.Labels[core.LabelReaper] = "true"
	req.Labels[core.LabelRyuk] = "true"
	delete(req.Labels, core.LabelReap)

	for _, customize := range customize {
		customize(&req)
	}

	return req
}

// reaperDisable disables / enables the reaper for the duration of the test.
func reaperDisable(t *testing.T, disabled bool) {
	t.Helper()

	config.Reset()
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", strconv.FormatBool(disabled))
	t.Cleanup(config.Reset)
}

func testContainerStart(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	ctr, err := Run(ctx, nginxAlpineImage, WithExposedPorts(nginxDefaultPort))
	CleanupContainer(t, ctr)
	require.NoError(t, err)
}

// testReaperRunning validates that a reaper is running.
func testReaperRunning(t *testing.T) {
	t.Helper()

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	ctx := context.Background()
	sessionID := config.Read().SessionID
	reaperContainer, err := spawner.lookupContainer(ctx, sessionID, provider)
	require.NoError(t, err)
	require.NotNil(t, reaperContainer)
}

func TestContainer(t *testing.T) {
	reaperDisable(t, false)

	t.Run("start/reaper-enabled", func(t *testing.T) {
		testContainerStart(t)
		testReaperRunning(t)
	})

	t.Run("stop/reaper-enabled", func(t *testing.T) {
		testContainerStop(t)
		testReaperRunning(t)
	})

	t.Run("terminate/reaper-enabled", func(t *testing.T) {
		testContainerTerminate(t)
		testReaperRunning(t)
	})

	reaperDisable(t, true)

	t.Run("start/reaper-disabled", func(t *testing.T) {
		testContainerStart(t)
	})

	t.Run("stop/reaper-disabled", func(t *testing.T) {
		testContainerStop(t)
	})

	t.Run("terminate/reaper-disabled", func(t *testing.T) {
		testContainerTerminate(t)
	})
}

// testContainerStop tests stopping a container.
func testContainerStop(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	nginxA, err := Run(ctx, nginxAlpineImage, WithExposedPorts(nginxDefaultPort))
	CleanupContainer(t, nginxA)
	require.NoError(t, err)

	state, err := nginxA.State(ctx)
	require.NoError(t, err)
	require.True(t, state.Running)

	stopTimeout := 10 * time.Second
	err = nginxA.Stop(ctx, &stopTimeout)
	require.NoError(t, err)

	state, err = nginxA.State(ctx)
	require.NoError(t, err)
	require.False(t, state.Running)
	require.Equal(t, container.StateExited, state.Status)
}

// testContainerTerminate tests terminating a container.
func testContainerTerminate(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	nginxA, err := Run(ctx, nginxAlpineImage, WithExposedPorts(nginxDefaultPort))
	CleanupContainer(t, nginxA)
	require.NoError(t, err)

	state, err := nginxA.State(ctx)
	require.NoError(t, err)
	require.True(t, state.Running)

	err = nginxA.Terminate(ctx)
	require.NoError(t, err)

	_, err = nginxA.State(ctx)
	require.Error(t, err)
}

func Test_NewReaper(t *testing.T) {
	reaperDisable(t, false)

	ctx := context.Background()

	t.Run("non-privileged", func(t *testing.T) {
		testNewReaper(ctx, t,
			config.Config{
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Second,
			},
			expectedReaperRequest(),
			false,
		)
	})

	t.Run("privileged", func(t *testing.T) {
		testNewReaper(ctx, t,
			config.Config{
				RyukPrivileged:          true,
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Second,
			},
			expectedReaperRequest(),
			true,
		)
	})

	t.Run("custom-timeouts", func(t *testing.T) {
		testNewReaper(ctx, t,
			config.Config{
				RyukPrivileged:          true,
				RyukConnectionTimeout:   2 * time.Minute,
				RyukReconnectionTimeout: 20 * time.Second,
			},
			expectedReaperRequest(func(req *ContainerRequest) {
				req.Env = map[string]string{
					"RYUK_PORT":                 "8080",
					"RYUK_CONNECTION_TIMEOUT":   "2m0s",
					"RYUK_RECONNECTION_TIMEOUT": "20s",
				}
			}),
			true,
		)
	})

	t.Run("verbose", func(t *testing.T) {
		testNewReaper(ctx, t,
			config.Config{
				RyukPrivileged: true,
				RyukVerbose:    true,
			},
			expectedReaperRequest(func(req *ContainerRequest) {
				req.Env = map[string]string{
					"RYUK_PORT":    "8080",
					"RYUK_VERBOSE": "true",
				}
			}),
			true,
		)
	})

	t.Run("docker-host", func(t *testing.T) {
		testNewReaper(context.WithValue(ctx, core.DockerHostContextKey, core.DockerSocketPathWithSchema), t,
			config.Config{
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Second,
			},
			expectedReaperRequest(func(req *ContainerRequest) {
				req.HostConfigModifier = func(hostConfig *container.HostConfig) {
					hostConfig.Binds = []string{core.MustExtractDockerSocket(ctx) + ":/var/run/docker.sock"}
				}
			}),
			false,
		)
	})

	t.Run("hub-prefix", func(t *testing.T) {
		testNewReaper(context.WithValue(ctx, core.DockerHostContextKey, core.DockerSocketPathWithSchema), t,
			config.Config{
				HubImageNamePrefix:      "registry.mycompany.com/mirror",
				RyukPrivileged:          true,
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Second,
			},
			expectedReaperRequest(func(req *ContainerRequest) {
				req.Image = config.ReaperDefaultImage
				req.HostConfigModifier = func(hc *container.HostConfig) {
					hc.Privileged = true
				}
			}),
			true,
		)
	})

	t.Run("hub-prefix-env", func(t *testing.T) {
		config.Reset()
		t.Cleanup(config.Reset)

		t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "registry.mycompany.com/mirror")
		testNewReaper(context.WithValue(ctx, core.DockerHostContextKey, core.DockerSocketPathWithSchema), t,
			config.Config{
				RyukPrivileged:          true,
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Second,
			},
			expectedReaperRequest(func(req *ContainerRequest) {
				req.Image = config.ReaperDefaultImage
				req.HostConfigModifier = func(hc *container.HostConfig) {
					hc.Privileged = true
				}
			}),
			true,
		)
	})
}

func testNewReaper(ctx context.Context, t *testing.T, cfg config.Config, expected ContainerRequest, privileged bool) {
	t.Helper()

	if prefix := os.Getenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX"); prefix != "" {
		cfg.HubImageNamePrefix = prefix
	}

	provider := newMockReaperProvider(cfg)

	// We need a new reaperSpawner for each test case to avoid reusing
	// an existing reaper instance.
	spawner := newReaperSpawner()
	reaper, err := spawner.reaper(ctx, testSessionID, provider)
	cleanupReaper(t, reaper, spawner)
	// We should have errored out see mockReaperProvider.RunContainer.
	require.ErrorIs(t, err, errExpected)

	require.Equal(t, expected.Image, provider.req.Image, "expected image doesn't match the submitted request")
	require.Equal(t, expected.ExposedPorts, provider.req.ExposedPorts, "expected exposed ports don't match the submitted request")
	require.Equal(t, expected.Labels, provider.req.Labels, "expected labels don't match the submitted request")
	require.Equal(t, expected.Mounts, provider.req.Mounts, "expected mounts don't match the submitted request")
	require.Equal(t, expected.WaitingFor, provider.req.WaitingFor, "expected waitingFor don't match the submitted request")
	require.Equal(t, expected.Env, provider.req.Env, "expected env doesn't match the submitted request")

	// checks for reaper's preCreationCallback fields
	require.Equal(t, container.NetworkMode(Bridge), provider.hostConfig.NetworkMode, "expected networkMode doesn't match the submitted request")
	require.True(t, provider.hostConfig.AutoRemove, "expected networkMode doesn't match the submitted request")
	require.Equal(t, privileged, provider.hostConfig.Privileged, "expected privileged doesn't match the submitted request")
}

func Test_ReaperReusedIfHealthy(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()
	// As other integration tests run with the (shared) Reaper as well, re-use the instance to not interrupt other tests
	if spawner.instance != nil {
		t.Cleanup(func() {
			require.NoError(t, spawner.cleanup())
		})
	}

	provider, err := ProviderDefault.GetProvider()
	require.NoError(t, err)

	reaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.(*DockerProvider).host), testSessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")

	reaperReused, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.(*DockerProvider).host), testSessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "reusing the Reaper should not error")

	// Ensure the internal state of both reaper instances is the same
	require.Equal(t, reaper.SessionID, reaperReused.SessionID, "expecting the same SessionID")
	require.Equal(t, reaper.Endpoint, reaperReused.Endpoint, "expecting the same reaper endpoint")
	require.Equal(t, reaper.Provider, reaperReused.Provider, "expecting the same container provider")
	require.Equal(t, reaper.container.GetContainerID(), reaperReused.container.GetContainerID(), "expecting the same container ID")
	require.Equal(t, reaper.container.SessionID(), reaperReused.container.SessionID(), "expecting the same session ID")

	termSignal, err := reaper.Connect()
	cleanupTermSignal(t, termSignal)
	require.NoError(t, err, "connecting to Reaper should be successful")
}

func Test_RecreateReaperIfTerminated(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := ProviderDefault.GetProvider()
	require.NoError(t, err)

	ctx := context.Background()
	reaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.(*DockerProvider).host), testSessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")

	termSignal, err := reaper.Connect()
	if termSignal != nil {
		termSignal <- true
	}
	require.NoError(t, err)

	// Wait for up to ryuk's default reconnect timeout + 1s to allow for a graceful shutdown/cleanup of the container.
	timeout := time.NewTimer(time.Second * 11)
	t.Cleanup(func() {
		timeout.Stop()
	})
	for {
		// Listing also handles auto-removal without racing an inspect against
		// Podman's deletion of the container's state from its database.
		containers, err := provider.(*DockerProvider).client.ContainerList(ctx, client.ContainerListOptions{
			All:     true,
			Filters: make(client.Filters).Add("id", reaper.container.GetContainerID()),
		})
		require.NoError(t, err)
		if len(containers.Items) == 0 || containers.Items[0].State != container.StateRunning {
			break
		}

		select {
		case <-timeout.C:
			t.Fatal("reaper container should have been terminated")
		default:
		}

		time.Sleep(time.Millisecond * 100)
	}

	recreatedReaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.(*DockerProvider).host), testSessionID, provider)
	cleanupReaper(t, recreatedReaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")
	require.NotEqual(t, reaper.container.GetContainerID(), recreatedReaper.container.GetContainerID(), "expected different container ID")

	recreatedTermSignal, err := recreatedReaper.Connect()
	cleanupTermSignal(t, recreatedTermSignal)
	require.NoError(t, err, "connecting to Reaper should be successful")
}

// Test_RecreateReaperIfStopped tests that a reaper container which still exists
// but is no longer running, e.g. it shut down after its reconnection timeout
// with no clients but was not removed yet, is replaced instead of being waited
// on until the startup timeout expires.
func Test_RecreateReaperIfStopped(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()

	provider, err := NewDockerProvider()
	require.NoError(t, err)

	// Create a stopped container that the lookup identifies as the session's
	// reaper: same name and labels, but exited and not auto-removed.
	id := startReaperTestContainer(ctx, t, provider, true, "true")
	cli := provider.Client()
	requireContainerStopped(ctx, t, cli, id)

	// The stopped container must be replaced by a fresh reaper well within
	// the startup timeout that waiting on it for readiness would burn.
	timeout, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	spawner := newReaperSpawner()
	reaper, err := spawner.reaper(context.WithValue(timeout, core.DockerHostContextKey, provider.host), testSessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")
	require.NotEqual(t, id, reaper.container.GetContainerID(), "expected a new reaper container")

	// The stopped container was removed to free up the reaper name.
	_, err = cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
	require.True(t, errdefs.IsNotFound(err), "stopped reaper container should have been removed, got: %v", err)

	termSignal, err := reaper.Connect()
	cleanupTermSignal(t, termSignal)
	require.NoError(t, err, "connecting to Reaper should be successful")
}

// Test_LookupReaperKeepsPausedContainer tests that a paused reaper is not
// removed: it can come back, and other test processes may still depend on it.
func Test_LookupReaperKeepsPausedContainer(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()

	provider, err := NewDockerProvider()
	require.NoError(t, err)

	id := startReaperTestContainer(ctx, t, provider, true, "sleep", "300")
	cli := provider.Client()
	_, err = cli.ContainerPause(ctx, id, client.ContainerPauseOptions{})
	require.NoError(t, err)

	timeout, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	spawner := newReaperSpawner()
	_, err = spawner.lookupContainer(context.WithValue(timeout, core.DockerHostContextKey, provider.host), testSessionID, provider)
	require.Error(t, err, "a paused reaper should not be reused")
	require.NotErrorIs(t, err, errReaperNotFound, "a paused reaper should not be replaced")

	inspect, err := cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
	require.NoError(t, err, "paused reaper container should not have been removed")
	require.True(t, inspect.Container.State.Paused, "reaper container should still be paused")
}

// Test_ReaperFromContainerStopped tests that waiting on a reaper which stops
// before or during the readiness wait fails quickly with a retryable error,
// so the spawner recreates it instead of waiting for the startup timeout.
func Test_ReaperFromContainerStopped(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)

	// requireRetryable requires that reusing container id fails quickly with
	// a retryable not found error.
	requireRetryable := func(ctx context.Context, t *testing.T, spawner *reaperSpawner, id string) {
		t.Helper()

		ctr := &DockerContainer{
			ID:       id,
			Image:    alpineImage,
			provider: provider,
			logger:   provider.Logger,
		}

		start := time.Now()
		_, err := spawner.fromContainer(ctx, testSessionID, provider, ctr)
		require.Less(t, time.Since(start), time.Second*30, "should fail well before the startup timeout")
		require.True(t, errdefs.IsNotFound(err), "expected a not found error, got: %v", err)

		var permanent *backoff.PermanentError
		require.NotErrorAs(t, spawner.retryError(err), &permanent, "error should be retryable")
	}

	t.Run("before-wait", func(t *testing.T) {
		ctx := context.Background()
		id := startReaperTestContainer(ctx, t, provider, false, "true")
		requireContainerStopped(ctx, t, provider.Client(), id)

		requireRetryable(ctx, t, newReaperSpawner(), id)
	})

	t.Run("during-wait", func(t *testing.T) {
		ctx := context.Background()
		id := startReaperTestContainer(ctx, t, provider, false, "sleep", "300")

		// Stop the container from inside the wait, so it is stopped for the
		// check after the wait, whatever the timing.
		spawner := newReaperSpawner()
		spawner.waitStrategy = func(string, *reaperConn) wait.Strategy {
			return stopOnWait{cli: provider.Client(), id: id}
		}
		requireRetryable(ctx, t, spawner, id)
	})
}

// inspectCounter is an APIClient that counts its container inspects.
type inspectCounter struct {
	client.APIClient
	inspects atomic.Int32
}

func (c *inspectCounter) ContainerInspect(ctx context.Context, id string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	c.inspects.Add(1)
	return c.APIClient.ContainerInspect(ctx, id, options)
}

// Test_ReaperFromContainerWaitsFirst tests that reusing a reaper waits on it
// before anything else, as one with no clients is counting down its
// reconnection timeout until it's connected to.
func Test_ReaperFromContainerWaitsFirst(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	sessionID := testSessionID + "-waits-first"
	t.Cleanup(func() {
		_, err := provider.Client().ContainerRemove(context.Background(), reaperContainerNameFromSessionID(sessionID), client.ContainerRemoveOptions{Force: true})
		require.True(t, isCleanupSafe(err), "remove reaper: %v", err)
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	creator := newReaperSpawner()
	created, err := creator.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.host), sessionID, provider)
	cleanupReaper(t, created, creator)
	require.NoError(t, err, "creating the Reaper should not error")

	reuser, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reuser.Close()) })

	counter := &inspectCounter{APIClient: reuser.Client()}
	reuser.SetClient(counter)

	ctr := &DockerContainer{
		ID:       created.container.GetContainerID(),
		Image:    config.ReaperDefaultImage,
		provider: reuser,
		logger:   reuser.Logger,
	}

	spawner := newReaperSpawner()
	inspectsBefore := int32(-1)
	spawner.waitStrategy = func(sessionID string, conn *reaperConn) wait.Strategy {
		inspectsBefore = counter.inspects.Load()
		return spawner.defaultWaitStrategy(sessionID, conn)
	}

	reused, err := spawner.fromContainer(ctx, sessionID, reuser, ctr)
	require.NoError(t, err)
	reused.close()
	require.Zero(t, inspectsBefore, "nothing should be inspected before the reaper is waited on")
}

// stopOnWait is a wait strategy that stops the container and fails, to
// simulate a reaper that stops while it is being waited on.
type stopOnWait struct {
	cli client.APIClient
	id  string
}

func (s stopOnWait) WaitUntilReady(ctx context.Context, _ wait.StrategyTarget) error {
	timeout := 0
	if _, err := s.cli.ContainerStop(ctx, s.id, client.ContainerStopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("stop: %w", err)
	}

	return errors.New("not ready")
}

// notReadyOnce records the containers waited for by the strategies it wraps,
// and fails the first wait, like one waiting on a reaper whose port Docker
// never publishes.
type notReadyOnce struct {
	mtx sync.Mutex
	ids []string
}

// wrap returns a wait strategy that records the container it waits for, then
// unless it's the first, waits with ready.
func (s *notReadyOnce) wrap(ready wait.Strategy) wait.Strategy {
	return notReadyOnceWait{once: s, ready: ready}
}

func (s *notReadyOnce) IDs() []string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return slices.Clone(s.ids)
}

// notReadyOnceWait is the wait strategy returned by notReadyOnce.wrap.
type notReadyOnceWait struct {
	once  *notReadyOnce
	ready wait.Strategy
}

func (s notReadyOnceWait) WaitUntilReady(ctx context.Context, target wait.StrategyTarget) error {
	ctr, ok := target.(interface{ GetContainerID() string })
	if !ok {
		return fmt.Errorf("unexpected target %T", target)
	}

	s.once.mtx.Lock()
	s.once.ids = append(s.once.ids, ctr.GetContainerID())
	first := len(s.once.ids) == 1
	s.once.mtx.Unlock()

	if first {
		<-ctx.Done()
		return ctx.Err()
	}

	return s.ready.WaitUntilReady(ctx, target)
}

// Test_ReaperRecreatedWhenNotReady tests that a new reaper which doesn't
// become ready within the ready timeout, e.g. because Docker didn't publish
// its port, see docker/desktop-feedback#338, is removed and replaced rather
// than waited on for the full startup timeout.
func Test_ReaperRecreatedWhenNotReady(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	sessionID := testSessionID + "-not-ready"
	t.Cleanup(func() {
		_, err := provider.Client().ContainerRemove(context.Background(), reaperContainerNameFromSessionID(sessionID), client.ContainerRemoveOptions{Force: true})
		require.True(t, isCleanupSafe(err), "remove reaper: %v", err)
	})

	// The second attempt waits on a real reaper with the same timeout, so it
	// mustn't be so short that a slow start fails it too.
	spawner := newReaperSpawner()
	spawner.readyTimeout = time.Second * 3
	strategy := &notReadyOnce{}
	spawner.waitStrategy = func(sessionID string, conn *reaperConn) wait.Strategy {
		return wait.ForAll(strategy.wrap(spawner.defaultWaitStrategy(sessionID, conn))).WithDeadline(spawner.readyTimeout)
	}

	// Bound the test, so it fails rather than waits for the run timeout if
	// the first attempt isn't limited.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()
	reaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.host), sessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")
	require.Less(t, time.Since(start), time.Second*15, "the reaper should be replaced well before the startup timeout")

	ids := strategy.IDs()
	require.GreaterOrEqual(t, len(ids), 2, "a second reaper should have been created")
	require.NotEqual(t, ids[0], reaper.container.GetContainerID(), "expected a new reaper container")

	require.Eventually(t, func() bool {
		_, err := provider.Client().ContainerInspect(context.Background(), ids[0], client.ContainerInspectOptions{})
		return errdefs.IsNotFound(err)
	}, time.Second*10, time.Millisecond*100, "the reaper that wasn't ready should have been removed")

	termSignal, err := reaper.Connect()
	cleanupTermSignal(t, termSignal)
	require.NoError(t, err, "connecting to Reaper should be successful")
}

// Test_ReaperFromContainerNotReady tests that waiting on an existing reaper
// which never becomes ready gives up after the ready timeout, with a
// retryable error, rather than waiting for the full startup timeout. Its port
// is published, so other processes may be using it, and it is kept.
func Test_ReaperFromContainerNotReady(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)

	ctx := context.Background()
	id := startUnreadyReaperContainer(ctx, t, provider, true)
	ctr := &DockerContainer{
		ID:       id,
		Image:    alpineImage,
		provider: provider,
		logger:   provider.Logger,
	}

	spawner := newReaperSpawner()
	spawner.readyTimeout = time.Second
	errs := make(chan error, 1)
	start := time.Now()
	go func() {
		_, err := spawner.fromContainer(ctx, testSessionID, provider, ctr)
		errs <- err
	}()

	select {
	case err := <-errs:
		require.Less(t, time.Since(start), time.Second*5, "should give up once the ready timeout has passed")
		require.ErrorIs(t, err, context.DeadlineExceeded)

		var permanent *backoff.PermanentError
		require.NotErrorAs(t, spawner.retryError(err), &permanent, "error should be retryable")
	case <-time.After(time.Second * 10):
		t.Fatal("fromContainer should give up once the ready timeout has passed")
	}

	_, err = provider.Client().ContainerInspect(ctx, id, client.ContainerInspectOptions{})
	require.NoError(t, err, "a reaper with a published port should not be removed")
}

// Test_ReaperReplacedWhenPortNotPublished tests that an existing reaper whose
// port Docker never published, see docker/desktop-feedback#338, is replaced:
// nothing can reach it, so nothing depends on it, and waiting on it again
// would never succeed.
func Test_ReaperReplacedWhenPortNotPublished(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)

	ctx := context.Background()
	id := startUnreadyReaperContainer(ctx, t, provider, false)

	timeout, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	spawner := newReaperSpawner()
	// Long enough to show the reaper is replaced without waiting for it.
	spawner.readyTimeout = time.Minute
	start := time.Now()
	reaper, err := spawner.reaper(context.WithValue(timeout, core.DockerHostContextKey, provider.host), testSessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")
	require.Less(t, time.Since(start), time.Second*10, "the reaper should be replaced without waiting for the ready timeout, for it to stop or for the retries")
	require.NotEqual(t, id, reaper.container.GetContainerID(), "expected a new reaper container")

	_, err = provider.Client().ContainerInspect(ctx, id, client.ContainerInspectOptions{})
	require.True(t, errdefs.IsNotFound(err), "the reaper with no published port should have been removed, got: %v", err)

	termSignal, err := reaper.Connect()
	cleanupTermSignal(t, termSignal)
	require.NoError(t, err, "connecting to Reaper should be successful")
}

// createRecorder is an APIClient that records the names of the containers it
// creates.
type createRecorder struct {
	client.APIClient
	mtx   sync.Mutex
	names []string
}

func (c *createRecorder) ContainerCreate(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
	c.mtx.Lock()
	c.names = append(c.names, options.Name)
	c.mtx.Unlock()

	return c.APIClient.ContainerCreate(ctx, options)
}

func (c *createRecorder) Names() []string {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return slices.Clone(c.names)
}

// Test_NewReaperUsesProviderClient tests that a reaper is created with the
// provider's client, which may have been set with SetClient, so it runs on
// the same daemon as the provider's other containers.
func Test_NewReaperUsesProviderClient(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	recorder := &createRecorder{APIClient: provider.Client()}
	provider.SetClient(recorder)

	sessionID := testSessionID + "-provider-client"
	t.Cleanup(func() {
		_, err := recorder.ContainerRemove(context.Background(), reaperContainerNameFromSessionID(sessionID), client.ContainerRemoveOptions{Force: true})
		require.True(t, isCleanupSafe(err), "remove reaper: %v", err)
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	spawner := newReaperSpawner()
	reaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.host), sessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")
	require.Contains(t, recorder.Names(), reaperContainerNameFromSessionID(sessionID), "the reaper should be created with the provider's client")
}

// Test_NewReaperConnectsOnce tests that creating a reaper only connects to it
// once, keeping the connection it was acknowledged on, rather than also
// waiting on another connection's handshake.
func Test_NewReaperConnectsOnce(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	sessionID := testSessionID + "-connects-once"
	t.Cleanup(func() {
		_, err := provider.Client().ContainerRemove(context.Background(), reaperContainerNameFromSessionID(sessionID), client.ContainerRemoveOptions{Force: true})
		require.True(t, isCleanupSafe(err), "remove reaper: %v", err)
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	spawner := newReaperSpawner()
	reaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.host), sessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")

	logs, err := reaper.container.Logs(ctx)
	require.NoError(t, err)
	defer logs.Close()

	out, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(string(out), `msg="client connected"`), "the reaper should only be connected to once:\n%s", out)
}

// hidePortsOnce is an APIClient that reports no published ports for the first
// reaper container it inspects, like Docker in docker/desktop-feedback#338.
type hidePortsOnce struct {
	client.APIClient
	mtx sync.Mutex
	id  string
}

func (c *hidePortsOnce) ContainerInspect(ctx context.Context, id string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	res, err := c.APIClient.ContainerInspect(ctx, id, options)
	if err != nil || res.Container.Config == nil || !isReaperImage(res.Container.Config.Image) {
		return res, err
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.id == "" {
		c.id = res.Container.ID
	}

	if c.id == res.Container.ID && res.Container.NetworkSettings != nil {
		res.Container.NetworkSettings.Ports = network.PortMap{}
	}

	return res, nil
}

// Test_NewReaperReplacedWhenPortNotPublished tests that a new reaper whose
// port Docker didn't publish is replaced without waiting for it to stop, as
// nothing can be connected to it.
func Test_NewReaperReplacedWhenPortNotPublished(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	hide := &hidePortsOnce{APIClient: provider.Client()}
	provider.SetClient(hide)

	sessionID := testSessionID + "-new-not-published"
	t.Cleanup(func() {
		_, err := provider.Client().ContainerRemove(context.Background(), reaperContainerNameFromSessionID(sessionID), client.ContainerRemoveOptions{Force: true})
		require.True(t, isCleanupSafe(err), "remove reaper: %v", err)
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	spawner := newReaperSpawner()
	start := time.Now()
	reaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.host), sessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")
	require.Less(t, time.Since(start), time.Second*5, "the reaper should be replaced without waiting for it to stop")
	require.NotEmpty(t, hide.id, "a reaper's ports should have been hidden")
	require.NotEqual(t, hide.id, reaper.container.GetContainerID(), "expected a new reaper container")
}

// networkListCounter is an APIClient that counts its network lists.
type networkListCounter struct {
	client.APIClient
	lists atomic.Int32
}

func (c *networkListCounter) NetworkList(ctx context.Context, options client.NetworkListOptions) (client.NetworkListResult, error) {
	c.lists.Add(1)
	return c.APIClient.NetworkList(ctx, options)
}

// Test_NewReaperUsesProviderNetwork tests that a reaper is created on the
// default network its provider has found, without finding it again.
func Test_NewReaperUsesProviderNetwork(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	counter := &networkListCounter{APIClient: provider.Client()}
	provider.SetClient(counter)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, err = provider.ensureDefaultNetwork(ctx)
	require.NoError(t, err)
	counter.lists.Store(0)

	sessionID := testSessionID + "-provider-network"
	t.Cleanup(func() {
		_, err := provider.Client().ContainerRemove(context.Background(), reaperContainerNameFromSessionID(sessionID), client.ContainerRemoveOptions{Force: true})
		require.True(t, isCleanupSafe(err), "remove reaper: %v", err)
	})

	spawner := newReaperSpawner()
	reaper, err := spawner.newReaper(context.WithValue(ctx, core.DockerHostContextKey, provider.host), sessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")
	require.Zero(t, counter.lists.Load(), "the provider's default network should be used")
}

// Test_NewReaperUsesProviderHost tests that a reaper is reached at the host
// its provider has resolved, as the provider's other containers are, even if
// the host would now resolve differently.
func Test_NewReaperUsesProviderHost(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	t.Setenv("TESTCONTAINERS_HOST_OVERRIDE", "127.0.0.1")
	host, err := provider.DaemonHost(context.Background())
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", host)

	// An address nothing listens on, see RFC 5737.
	t.Setenv("TESTCONTAINERS_HOST_OVERRIDE", "203.0.113.1")

	sessionID := testSessionID + "-provider-host"
	t.Cleanup(func() {
		_, err := provider.Client().ContainerRemove(context.Background(), reaperContainerNameFromSessionID(sessionID), client.ContainerRemoveOptions{Force: true})
		require.True(t, isCleanupSafe(err), "remove reaper: %v", err)
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	spawner := newReaperSpawner()
	spawner.readyTimeout = time.Second * 3
	reaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.host), sessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")

	endpointHost, _, err := net.SplitHostPort(reaper.Endpoint)
	require.NoError(t, err)
	require.Equal(t, host, endpointHost, "the reaper should be reached at the provider's host")
}

// removedContainer is a Container that Podman removed while it was being
// inspected, so its inspects fail with an internal error rather than not
// found.
type removedContainer struct {
	Container
}

func (removedContainer) GetContainerID() string {
	return "removed-container"
}

func (removedContainer) State(context.Context) (*container.State, error) {
	return nil, errPodmanRemovedDuringInspect
}

func (removedContainer) Terminate(context.Context, ...TerminateOption) error {
	return errPodmanRemovedDuringInspect
}

// Test_ReaperRecreatedWhenRemovedDuringInspect tests that a reaper that
// Podman removes while it's being inspected, e.g. automatically once it
// exits, is recreated rather than failing permanently.
func Test_ReaperRecreatedWhenRemovedDuringInspect(t *testing.T) {
	reaperDisable(t, false)

	spawner := newReaperSpawner()
	spawner.instance = &Reaper{SessionID: testSessionID, container: removedContainer{}}

	_, err := spawner.retryLocked(context.Background(), testSessionID, newMockReaperProvider(config.Config{}))()
	require.ErrorIs(t, err, errPodmanRemovedDuringInspect)

	var permanent *backoff.PermanentError
	require.NotErrorAs(t, err, &permanent, "error should be retryable")
	require.Nil(t, spawner.instance, "the removed reaper should be forgotten")
}

// slowInspectClient is an APIClient whose container inspects take delay, like
// those of a slow daemon.
type slowInspectClient struct {
	client.APIClient
	delay time.Duration
}

func (c *slowInspectClient) ContainerInspect(ctx context.Context, id string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	select {
	case <-ctx.Done():
		return client.ContainerInspectResult{}, ctx.Err()
	case <-time.After(c.delay):
	}

	return c.APIClient.ContainerInspect(ctx, id, options)
}

// Test_NewReaperSlowDaemon tests that a new reaper stays running while it is
// connected to, even if that takes longer than its reconnection timeout, as
// it may on a slow daemon.
func Test_NewReaperSlowDaemon(t *testing.T) {
	reaperDisable(t, false)
	t.Setenv("TESTCONTAINERS_RYUK_RECONNECTION_TIMEOUT", "1s")

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	provider.SetClient(&slowInspectClient{APIClient: provider.Client(), delay: time.Millisecond * 1500})

	sessionID := testSessionID + "-slow-daemon"
	t.Cleanup(func() {
		_, err := provider.Client().ContainerRemove(context.Background(), reaperContainerNameFromSessionID(sessionID), client.ContainerRemoveOptions{Force: true})
		require.True(t, isCleanupSafe(err), "remove reaper: %v", err)
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	spawner := newReaperSpawner()
	reaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.host), sessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")
}

// Test_ReaperCreatedNotStarted tests a reaper container that was created but
// not started, e.g. because a caller deadline cancelled the create request
// after the daemon had already completed it. A recent one may still be
// started by the process that created it, so it is kept, but one that has
// stayed created past the grace period is an orphan and is replaced.
func Test_ReaperCreatedNotStarted(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	provider, err := NewDockerProvider()
	require.NoError(t, err)

	cli := provider.Client()

	t.Run("recent-is-kept", func(t *testing.T) {
		ctx := context.Background()
		id := createReaperTestContainer(ctx, t, provider, true, "sleep", "300")

		timeout, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()

		spawner := newReaperSpawner()
		_, err := spawner.lookupContainer(context.WithValue(timeout, core.DockerHostContextKey, provider.host), testSessionID, provider)
		require.Error(t, err, "a created reaper should not be reused")
		require.NotErrorIs(t, err, errReaperNotFound, "a recently created reaper should not be replaced")

		inspect, err := cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
		require.NoError(t, err, "recently created reaper container should not have been removed")
		require.Equal(t, container.StateCreated, inspect.Container.State.Status)
	})

	t.Run("orphan-is-replaced", func(t *testing.T) {
		ctx := context.Background()
		id := createReaperTestContainer(ctx, t, provider, true, "sleep", "300")

		timeout, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()

		// Shorten the grace period so the container counts as orphaned
		// straight away, rather than waiting for the default to pass.
		spawner := newReaperSpawner()
		spawner.createdGracePeriod = time.Nanosecond
		start := time.Now()
		reaper, err := spawner.reaper(context.WithValue(timeout, core.DockerHostContextKey, provider.host), testSessionID, provider)
		cleanupReaper(t, reaper, spawner)
		require.NoError(t, err, "creating the Reaper should not error")
		require.Less(t, time.Since(start), time.Second*15, "orphan should be replaced without waiting out the lookup retries")
		require.NotEqual(t, id, reaper.container.GetContainerID(), "expected a new reaper container")

		_, err = cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
		require.True(t, errdefs.IsNotFound(err), "orphaned reaper container should have been removed, got: %v", err)

		termSignal, err := reaper.Connect()
		cleanupTermSignal(t, termSignal)
		require.NoError(t, err, "connecting to Reaper should be successful")
	})
}

// Test_ReaperCheckStale tests which reaper containers are replaced. The age of
// a created container is measured with the daemon's clock, so the test uses a
// daemon time far from the client's clock to catch any use of the latter.
func Test_ReaperCheckStale(t *testing.T) {
	daemonTime := time.Unix(1_000_000_000, 0)
	created := func(age time.Duration) int64 {
		return daemonTime.Add(-age).Unix()
	}

	tests := []struct {
		name      string
		item      container.Summary
		daemonErr error
		stale     bool
	}{
		{name: "running", item: container.Summary{State: container.StateRunning}},
		{name: "exited", item: container.Summary{State: container.StateExited}, stale: true},
		{name: "dead", item: container.Summary{State: container.StateDead}, stale: true},
		{name: "removing", item: container.Summary{State: container.StateRemoving}, stale: true},
		{name: "paused", item: container.Summary{State: container.StatePaused}},
		{name: "restarting", item: container.Summary{State: container.StateRestarting}},
		{name: "unknown", item: container.Summary{State: ""}},
		{name: "created-recent", item: container.Summary{State: container.StateCreated, Created: created(time.Second * 5)}},
		{name: "created-orphan", item: container.Summary{State: container.StateCreated, Created: created(time.Second * 11)}, stale: true},
		{name: "created-unknown-time", item: container.Summary{State: container.StateCreated}},
		{name: "created-unknown-daemon-time", item: container.Summary{State: container.StateCreated, Created: created(time.Hour)}, daemonErr: errExpected},
	}

	spawner := newReaperSpawner()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var calls int
			daemonNow := func() (time.Time, error) {
				calls++
				return daemonTime, tc.daemonErr
			}

			err := spawner.checkStale(tc.item, daemonNow)
			switch {
			case tc.stale:
				require.NoError(t, err)
			case tc.daemonErr != nil:
				// Without the daemon's time the container's age is unknown,
				// so it must be kept rather than judged by the client's clock.
				require.ErrorIs(t, err, tc.daemonErr)
			default:
				require.Error(t, err)
			}

			if tc.item.State != container.StateCreated {
				require.Zero(t, calls, "the daemon time is only needed for created containers")
			}
		})
	}
}

// flakyInfoClient is a SystemAPIClient whose Info fails on its first call,
// then reports a daemon clock an hour ahead of the client's.
type flakyInfoClient struct {
	client.SystemAPIClient
	calls int
}

func (c *flakyInfoClient) Info(context.Context, client.InfoOptions) (client.SystemInfoResult, error) {
	c.calls++
	if c.calls == 1 {
		return client.SystemInfoResult{}, errExpected
	}

	var res client.SystemInfoResult
	res.Info.SystemTime = time.Now().Add(time.Hour).Format(time.RFC3339Nano)

	return res, nil
}

// Test_ReaperDaemonClock tests that the daemon's clock offset is cached once
// fetched, but a failure to fetch it is not, so a temporary error is retried
// on the next lookup attempt rather than kept for the rest of the lookup.
func Test_ReaperDaemonClock(t *testing.T) {
	cli := &flakyInfoClient{}
	daemonNow := daemonClock(context.Background(), cli)

	_, err := daemonNow()
	require.ErrorIs(t, err, errExpected)

	now, err := daemonNow()
	require.NoError(t, err, "a failure to fetch the offset should be retried")
	require.WithinDuration(t, time.Now().Add(time.Hour), now, time.Second)

	_, err = daemonNow()
	require.NoError(t, err)
	require.Equal(t, 2, cli.calls, "a fetched offset should be cached")
}

// Test_ReaperClockOffset tests reading the daemon's clock offset.
func Test_ReaperClockOffset(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		offset, err := clockOffset(time.Now().Add(time.Hour).Format(time.RFC3339Nano))
		require.NoError(t, err)
		require.InDelta(t, time.Hour, offset, float64(time.Second))
	})

	t.Run("empty", func(t *testing.T) {
		_, err := clockOffset("")
		require.Error(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := clockOffset("not a time")
		require.Error(t, err)
	})

	t.Run("daemon-unavailable", func(t *testing.T) {
		// Nothing listens on port 1, so the info request fails.
		apiClient, err := client.New(client.WithHost("tcp://127.0.0.1:1"))
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, apiClient.Close()) })

		_, err = daemonClockOffset(context.Background(), apiClient)
		require.Error(t, err)
	})
}

// Test_RemoveStaleReaper tests removing a stale reaper container. A reaper
// that has started since it was listed, e.g. a created reaper whose delayed
// start completed in the meantime, must not be removed, as other test
// processes may already be using it.
func Test_RemoveStaleReaper(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()

	provider, err := NewDockerProvider()
	require.NoError(t, err)

	cli := provider.Client()

	t.Run("created", func(t *testing.T) {
		id := createReaperTestContainer(ctx, t, provider, false, "sleep", "300")

		require.NoError(t, removeStaleReaper(ctx, cli, id))

		_, err := cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
		require.True(t, errdefs.IsNotFound(err), "created container should have been removed, got: %v", err)
	})

	t.Run("started-since-listed", func(t *testing.T) {
		// Listed while created, then started before it is removed.
		id := createReaperTestContainer(ctx, t, provider, false, "sleep", "300")
		_, err := cli.ContainerStart(ctx, id, client.ContainerStartOptions{})
		require.NoError(t, err)

		err = removeStaleReaper(ctx, cli, id)
		require.Error(t, err, "a started reaper should not be removed")

		var permanent *backoff.PermanentError
		require.NotErrorAs(t, err, &permanent, "the lookup should retry and see the reaper is running")

		inspect, err := cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
		require.NoError(t, err, "started reaper container should not have been removed")
		require.True(t, inspect.Container.State.Running, "reaper container should still be running")
	})
}

// blockingReaperProvider is a ReaperProvider whose RunContainer blocks until
// released, to control when a reaper run finishes relative to its caller.
type blockingReaperProvider struct {
	*mockReaperProvider
	started chan context.Context
	release chan error
	ctr     Container
}

func (p *blockingReaperProvider) RunContainer(ctx context.Context, _ ContainerRequest) (Container, error) {
	p.started <- ctx
	return p.ctr, <-p.release
}

// terminateRecorder is a Container that records when it is terminated.
type terminateRecorder struct {
	Container
	terminated chan struct{}
}

func (c *terminateRecorder) GetContainerID() string {
	return "terminate-recorder"
}

func (c *terminateRecorder) Terminate(context.Context, ...TerminateOption) error {
	close(c.terminated)
	return nil
}

// stuckContainer is a Container whose Terminate blocks until its context is
// done, like one on an unresponsive daemon.
type stuckContainer struct {
	Container
}

func (stuckContainer) GetContainerID() string {
	return "stuck-container"
}

func (stuckContainer) Terminate(ctx context.Context, _ ...TerminateOption) error {
	<-ctx.Done()
	return ctx.Err()
}

// failingContainer is a Container whose Terminate fails.
type failingContainer struct {
	Container
}

var errTerminate = errors.New("terminate failed")

func (failingContainer) GetContainerID() string {
	return "failing-container"
}

func (failingContainer) Terminate(context.Context, ...TerminateOption) error {
	return errTerminate
}

// Test_NewReaperOutlivesCaller tests that a caller giving up doesn't abandon
// creating or starting the reaper half way, which can leave an orphaned
// reaper that is created but never started, see #3918. The caller gets its
// error straight away, and the run cleans up after itself if it fails.
func Test_NewReaperOutlivesCaller(t *testing.T) {
	ctr := &terminateRecorder{terminated: make(chan struct{})}
	provider := &blockingReaperProvider{
		mockReaperProvider: newMockReaperProvider(config.Config{}),
		started:            make(chan context.Context, 1),
		release:            make(chan error, 1),
		ctr:                ctr,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errs := make(chan error, 1)
	go func() {
		_, err := newReaperSpawner().newReaper(ctx, testSessionID, provider)
		errs <- err
	}()

	var runCtx context.Context
	select {
	case runCtx = <-provider.started:
	case <-time.After(time.Second * 10):
		t.Fatal("reaper run should have started")
	}

	cancel()

	select {
	case err := <-errs:
		require.ErrorIs(t, err, context.Canceled, "the caller should get its cancellation")
	case <-time.After(time.Second * 5):
		t.Fatal("newReaper should return when its caller's context is done")
	}
	require.NoError(t, runCtx.Err(), "the reaper run should not be cancelled with its caller")

	// The run fails after its caller has gone, so the container it created
	// must still be removed rather than left behind.
	provider.release <- errExpected
	select {
	case <-ctr.terminated:
	case <-time.After(time.Second * 5):
		t.Fatal("container should be terminated when the abandoned run fails")
	}
}

// Test_RunReaperClosesConn tests that the connection a reaper acknowledged
// is closed if the run fails or its caller has returned, so the reaper isn't
// kept running by a connection nothing uses.
func Test_RunReaperClosesConn(t *testing.T) {
	// run runs a reaper, with the connection a reaper acknowledged held,
	// whose RunContainer returns released once started, cancelling the
	// caller's context first if cancel is true. It returns the error and
	// the other end of the connection.
	run := func(t *testing.T, released error, cancel bool) (*reaperConn, net.Conn, error) {
		t.Helper()

		provider := &blockingReaperProvider{
			mockReaperProvider: newMockReaperProvider(config.Config{}),
			started:            make(chan context.Context, 1),
			release:            make(chan error, 1),
			ctr:                &terminateRecorder{terminated: make(chan struct{})},
		}

		local, remote := net.Pipe()
		t.Cleanup(func() { require.NoError(t, remote.Close()) })

		held := &reaperConn{}
		held.set(local, "127.0.0.1:8080")
		t.Cleanup(func() { require.NoError(t, held.Close()) })

		ctx, cancelCtx := context.WithCancel(context.Background())
		defer cancelCtx()

		errs := make(chan error, 1)
		go func() {
			_, err := newReaperSpawner().runReaper(ctx, provider, ContainerRequest{}, held, &detachableLogger{logger: &recordingLogger{}}, time.Minute)
			errs <- err
		}()

		select {
		case <-provider.started:
		case <-time.After(time.Second * 10):
			t.Fatal("reaper run should have started")
		}

		if cancel {
			cancelCtx()
		} else {
			provider.release <- released
		}

		var err error
		select {
		case err = <-errs:
		case <-time.After(time.Second * 5):
			t.Fatal("runReaper should have returned")
		}

		if cancel {
			provider.release <- released
		}

		return held, remote, err
	}

	requireClosed := func(t *testing.T, remote net.Conn) {
		t.Helper()

		errs := make(chan error, 1)
		go func() {
			_, err := remote.Read(make([]byte, 1))
			errs <- err
		}()

		select {
		case err := <-errs:
			require.ErrorIs(t, err, io.EOF, "the connection should be closed")
		case <-time.After(time.Second * 5):
			t.Fatal("the connection should be closed")
		}
	}

	t.Run("succeeded", func(t *testing.T) {
		held, _, err := run(t, nil, false)
		require.NoError(t, err)

		conn, _ := held.take()
		require.NotNil(t, conn, "the connection should be kept for the caller")
		require.NoError(t, conn.Close())
	})

	t.Run("failed", func(t *testing.T) {
		_, remote, err := run(t, errExpected, false)
		require.ErrorIs(t, err, errExpected)
		requireClosed(t, remote)
	})

	t.Run("abandoned", func(t *testing.T) {
		_, remote, err := run(t, nil, true)
		require.ErrorIs(t, err, context.Canceled)
		requireClosed(t, remote)
	})
}

// Test_RunReaperBoundsCleanup tests that removing the container of a failed
// run is bounded, so an unresponsive daemon can't hold the run, and so its
// caller, open indefinitely.
func Test_RunReaperBoundsCleanup(t *testing.T) {
	provider := &blockingReaperProvider{
		mockReaperProvider: newMockReaperProvider(config.Config{}),
		started:            make(chan context.Context, 1),
		release:            make(chan error, 1),
		ctr:                stuckContainer{},
	}
	provider.release <- errExpected

	spawner := newReaperSpawner()
	spawner.cleanupTimeout = time.Millisecond * 100
	errs := make(chan error, 1)
	go func() {
		_, err := spawner.runReaper(context.Background(), provider, ContainerRequest{}, &reaperConn{}, &detachableLogger{}, time.Minute)
		errs <- err
	}()

	select {
	case err := <-errs:
		require.ErrorIs(t, err, errExpected)
		require.ErrorIs(t, err, context.DeadlineExceeded, "cleanup should time out")
	case <-time.After(time.Second * 5):
		t.Fatal("runReaper should return once cleanup times out")
	}
}

// stuckReaperProvider is a ReaperProvider whose RunContainer doesn't return
// until its context is done, like one on a daemon which never responds.
type stuckReaperProvider struct {
	*mockReaperProvider
}

func (stuckReaperProvider) RunContainer(ctx context.Context, _ ContainerRequest) (Container, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

// Test_RunReaperTimeout tests that a reaper run gives up after its timeout,
// even though it isn't cancelled with its caller.
func Test_RunReaperTimeout(t *testing.T) {
	provider := stuckReaperProvider{mockReaperProvider: newMockReaperProvider(config.Config{})}

	errs := make(chan error, 1)
	go func() {
		_, err := newReaperSpawner().runReaper(context.Background(), provider, ContainerRequest{}, &reaperConn{}, &detachableLogger{}, time.Millisecond*100)
		errs <- err
	}()

	select {
	case err := <-errs:
		require.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(time.Second * 5):
		t.Fatal("runReaper should give up after its timeout")
	}
}

// recordingLogger is a Logger that records what is logged to it.
type recordingLogger struct {
	mtx   sync.Mutex
	lines []string
}

func (l *recordingLogger) Printf(format string, v ...any) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	l.lines = append(l.lines, fmt.Sprintf(format, v...))
}

func (l *recordingLogger) Lines() []string {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	return slices.Clone(l.lines)
}

// loggingReaperProvider is a ReaperProvider whose RunContainer logs before it
// is released and after, to check which logger a run that outlives its
// caller logs to.
type loggingReaperProvider struct {
	*mockReaperProvider
	logger   log.Logger
	started  chan struct{}
	release  chan struct{}
	finished chan struct{}
}

func (p *loggingReaperProvider) RunContainer(context.Context, ContainerRequest) (Container, error) {
	p.logger.Printf("before the caller returned")
	close(p.started)
	<-p.release
	p.logger.Printf("reaper run logging after its caller returned")
	close(p.finished)
	return nil, errExpected
}

// Test_RunReaperDetachesLogger tests that a run that outlives its caller
// stops logging to the caller's logger when the caller returns, as that may
// be a test logger which can't be used once its test has completed.
func Test_RunReaperDetachesLogger(t *testing.T) {
	callerLogger := &recordingLogger{}
	backgroundLogger := &recordingLogger{}
	logger := &detachableLogger{logger: callerLogger, background: backgroundLogger}
	provider := &loggingReaperProvider{
		mockReaperProvider: newMockReaperProvider(config.Config{}),
		logger:             logger,
		started:            make(chan struct{}),
		release:            make(chan struct{}),
		finished:           make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errs := make(chan error, 1)
	go func() {
		_, err := newReaperSpawner().runReaper(ctx, provider, ContainerRequest{}, &reaperConn{}, logger, time.Minute)
		errs <- err
	}()

	select {
	case <-provider.started:
	case <-time.After(time.Second * 10):
		t.Fatal("reaper run should have started")
	}

	cancel()

	select {
	case err := <-errs:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second * 5):
		t.Fatal("runReaper should return when its caller's context is done")
	}

	close(provider.release)
	select {
	case <-provider.finished:
	case <-time.After(time.Second * 5):
		t.Fatal("reaper run should have finished")
	}

	require.Equal(t, []string{"before the caller returned"}, callerLogger.Lines(), "only logs before the caller returned should use its logger")

	// The run's failure is reported to the background logger after this.
	lines := backgroundLogger.Lines()
	require.NotEmpty(t, lines, "logs after the caller returned should use the background logger")
	require.Equal(t, "reaper run logging after its caller returned", lines[0])
}

// Test_NewReaperSpawner tests that a reaperSpawner starts with the default
// settings, including the package's spawner.
func Test_NewReaperSpawner(t *testing.T) {
	// The background logger is the default logger when the package was
	// initialised, so a test logger set as the default since isn't used by a
	// reaper run that outlives its test.
	defaultLogger := log.Default()
	testLogger := &recordingLogger{}
	log.SetDefault(testLogger)
	t.Cleanup(func() { log.SetDefault(defaultLogger) })

	for name, s := range map[string]*reaperSpawner{
		"new":     newReaperSpawner(),
		"package": spawner,
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, defaultReaperCreatedGracePeriod, s.createdGracePeriod)
			require.Equal(t, reaperCleanupTimeout, s.cleanupTimeout)
			require.Equal(t, reaperReadyTimeout, s.readyTimeout)
			conn := &reaperConn{}
			require.Equal(t, s.defaultWaitStrategy(testSessionID, conn), s.waitStrategy(testSessionID, conn))
			require.Equal(t, backgroundLogger, s.detachedLogger)
			require.NotEqual(t, testLogger, s.detachedLogger, "the current default logger should not be used")
		})
	}

	t.Run("ready-timeout", func(t *testing.T) {
		s := newReaperSpawner()
		s.readyTimeout = time.Second
		conn := &reaperConn{}
		require.NotEqual(t, newReaperSpawner().waitStrategy(testSessionID, conn), s.waitStrategy(testSessionID, conn), "the wait strategy should use the ready timeout")
	})
}

// Test_ReaperFilters tests that the filters sent to a reaper are the same
// each time, as a reaper registers each different line as a new filter.
func Test_ReaperFilters(t *testing.T) {
	filters := reaperFilters(testSessionID)
	for range 20 {
		require.Equal(t, filters, reaperFilters(testSessionID))
	}

	require.True(t, bytes.HasSuffix(filters, []byte("\n")), "filters should end with a newline")
	require.Contains(t, string(filters), "label="+core.LabelSessionID+"="+testSessionID)
}

// strategyTarget is a wait.StrategyTarget whose reaper listens on a local
// port, published unless unpublished is set, and whose state is running
// unless stopped is set. It only supports the calls the reaper wait strategy
// needs, and counts its inspects.
type strategyTarget struct {
	wait.StrategyTarget
	port        uint16
	stopped     bool
	unpublished bool
	inspects    atomic.Int32
}

func (s *strategyTarget) Inspect(context.Context) (*container.InspectResponse, error) {
	s.inspects.Add(1)

	state := &container.State{Status: container.StateRunning, Running: true}
	if s.stopped {
		state = &container.State{Status: container.StateExited}
	}

	hostPort := strconv.Itoa(int(s.port))
	if s.unpublished {
		hostPort = ""
	}

	return &container.InspectResponse{
		State: state,
		NetworkSettings: &container.NetworkSettings{
			Ports: network.PortMap{
				network.MustParsePort("8080/tcp"): {{HostPort: hostPort}},
			},
		},
	}, nil
}

func (s *strategyTarget) Host(context.Context) (string, error) {
	return "127.0.0.1", nil
}

// serveTCP starts a local listener which handles each connection with handle,
// and returns its port.
func serveTCP(t *testing.T, handle func(conn net.Conn)) uint16 {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, ln.Close()) })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			go func() {
				defer conn.Close()
				handle(conn)
			}()
		}
	}()

	addr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)

	return uint16(addr.Port) //nolint:gosec // A TCP port fits in a uint16.
}

// Test_RyukHandshake tests that the reaper wait strategy waits for the
// reaper to acknowledge the session's filters, rather than just for its port
// to accept connections, which the docker-proxy does whether or not ryuk is
// listening.
func Test_RyukHandshake(t *testing.T) {
	// waitUntilReady waits for target and returns the connection kept, if any,
	// and its endpoint.
	waitUntilReady := func(t *testing.T, target wait.StrategyTarget) (net.Conn, string, error) {
		t.Helper()

		s := newReaperSpawner()
		s.readyTimeout = time.Millisecond * 500
		held := &reaperConn{}
		err := s.defaultWaitStrategy(testSessionID, held).WaitUntilReady(context.Background(), target)

		conn, endpoint := held.take()
		if conn != nil {
			t.Cleanup(func() { require.NoError(t, conn.Close()) })
		}

		return conn, endpoint, err
	}

	t.Run("ack", func(t *testing.T) {
		received := make(chan string, 1)
		closed := make(chan struct{})
		port := serveTCP(t, func(conn net.Conn) {
			r := bufio.NewReader(conn)
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}

			received <- line
			_, _ = conn.Write(reaperAck)

			_, _ = io.Copy(io.Discard, r)
			close(closed)
		})

		target := &strategyTarget{port: port}
		conn, endpoint, err := waitUntilReady(t, target)
		require.NoError(t, err)
		require.Equal(t, net.JoinHostPort("127.0.0.1", strconv.Itoa(int(port))), endpoint, "the endpoint acknowledged on should be kept")
		require.Equal(t, int32(1), target.inspects.Load(), "an attempt should inspect the container once")

		select {
		case line := <-received:
			require.Equal(t, string(reaperFilters(testSessionID)), line, "the session's filters should be sent")
		case <-time.After(time.Second * 5):
			t.Fatal("the session's filters should have been sent")
		}

		// The reaper starts its reconnection timeout when the connection is
		// closed, so it is kept for the Reaper.
		require.NotNil(t, conn, "the acknowledged connection should be kept")
		require.Never(t, func() bool {
			select {
			case <-closed:
				return true
			default:
				return false
			}
		}, time.Millisecond*200, time.Millisecond*20, "the acknowledged connection should not be closed")
	})

	t.Run("accepted-not-acknowledged", func(t *testing.T) {
		// Like the docker-proxy with nothing listening behind it.
		port := serveTCP(t, func(net.Conn) {})

		conn, _, err := waitUntilReady(t, &strategyTarget{port: port})
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Nil(t, conn, "no connection should be kept")
	})

	t.Run("no-response", func(t *testing.T) {
		port := serveTCP(t, func(conn net.Conn) {
			_, _ = io.Copy(io.Discard, conn)
		})

		conn, _, err := waitUntilReady(t, &strategyTarget{port: port})
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Nil(t, conn, "no connection should be kept")
	})

	t.Run("no-response-then-ack", func(t *testing.T) {
		// Each attempt gives up on a connection with no response, so a
		// later attempt can succeed within the ready timeout.
		var connections atomic.Int32
		port := serveTCP(t, func(conn net.Conn) {
			if connections.Add(1) == 1 {
				_, _ = io.Copy(io.Discard, conn)
				return
			}

			r := bufio.NewReader(conn)
			for {
				if _, err := r.ReadString('\n'); err != nil {
					return
				}

				_, _ = conn.Write(reaperAck)
			}
		})

		attemptTimeout := time.Millisecond * 100
		held := &reaperConn{}
		strategy := wait.ForAll(&ryukHandshake{
			filters:        reaperFilters(testSessionID),
			pollInterval:   time.Millisecond * 10,
			attemptTimeout: attemptTimeout,
			conn:           held,
		}).WithDeadline(time.Second * 2)
		require.NoError(t, strategy.WaitUntilReady(context.Background(), &strategyTarget{port: port}))
		require.GreaterOrEqual(t, connections.Load(), int32(2), "a second attempt should have been made")

		conn, _ := held.take()
		require.NotNil(t, conn, "the acknowledged connection should be kept")
		defer conn.Close()

		// The kept connection is the Reaper's, so mustn't time out with the
		// attempt.
		time.Sleep(attemptTimeout * 2)
		require.NoError(t, acknowledge(context.Background(), conn, reaperFilters(testSessionID)), "the kept connection should still be usable")
	})

	t.Run("unexpected-response", func(t *testing.T) {
		port := serveTCP(t, func(conn net.Conn) {
			_, _ = bufio.NewReader(conn).ReadString('\n')
			_, _ = conn.Write([]byte("NAK\n"))
		})

		conn, _, err := waitUntilReady(t, &strategyTarget{port: port})
		require.ErrorContains(t, err, "unexpected reaper response")
		require.Nil(t, conn, "no connection should be kept")
	})

	t.Run("stopped", func(t *testing.T) {
		var connections atomic.Int32
		port := serveTCP(t, func(net.Conn) {
			connections.Add(1)
		})

		start := time.Now()
		conn, _, err := waitUntilReady(t, &strategyTarget{port: port, stopped: true})
		require.Nil(t, conn, "no connection should be kept")
		require.Less(t, time.Since(start), time.Millisecond*400, "a stopped reaper should fail straight away")
		require.True(t, errdefs.IsNotFound(err), "expected a not found error, got: %v", err)
		require.Zero(t, connections.Load(), "a stopped reaper should not be connected to")
	})

	t.Run("not-published", func(t *testing.T) {
		start := time.Now()
		conn, _, err := waitUntilReady(t, &strategyTarget{unpublished: true})
		require.Nil(t, conn, "no connection should be kept")
		require.Less(t, time.Since(start), time.Millisecond*400, "a reaper whose port isn't published should fail straight away")
		require.ErrorIs(t, err, errReaperPortNotPublished)

		var permanent *backoff.PermanentError
		require.NotErrorAs(t, newReaperSpawner().retryError(err), &permanent, "error should be retryable")
	})
}

// Test_ReaperConnect tests that connecting to a reaper waits for it to
// acknowledge the session's filters, failing if it doesn't, rather than
// leaving a connection whose handshake may never complete.
func Test_ReaperConnect(t *testing.T) {
	connect := func(t *testing.T, handle func(conn net.Conn)) (chan bool, error) {
		t.Helper()

		port := serveTCP(t, handle)
		reaper := &Reaper{SessionID: testSessionID, Endpoint: net.JoinHostPort("127.0.0.1", strconv.Itoa(int(port)))}

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()

		return reaper.connect(ctx)
	}

	t.Run("ack", func(t *testing.T) {
		received := make(chan string, 1)
		closed := make(chan struct{})
		termSignal, err := connect(t, func(conn net.Conn) {
			r := bufio.NewReader(conn)
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}

			received <- line
			_, _ = conn.Write(reaperAck)

			_, _ = io.Copy(io.Discard, r)
			close(closed)
		})
		require.NoError(t, err)
		require.Equal(t, string(reaperFilters(testSessionID)), <-received, "the session's filters should be sent")

		termSignal <- true
		select {
		case <-closed:
		case <-time.After(time.Second * 5):
			t.Fatal("the connection should be closed when signalled")
		}
	})

	t.Run("no-response", func(t *testing.T) {
		start := time.Now()
		termSignal, err := connect(t, func(conn net.Conn) {
			_, _ = io.Copy(io.Discard, conn)
		})
		require.ErrorIs(t, err, os.ErrDeadlineExceeded)
		require.Nil(t, termSignal)
		require.Less(t, time.Since(start), time.Second*2, "should give up when ctx is done")
	})

	t.Run("unexpected-response", func(t *testing.T) {
		termSignal, err := connect(t, func(conn net.Conn) {
			_, _ = bufio.NewReader(conn).ReadString('\n')
			_, _ = conn.Write([]byte("NAK\n"))
		})
		require.ErrorContains(t, err, "unexpected reaper response")
		require.Nil(t, termSignal)
	})
}

// Test_ReaperCleanupClosesConn tests that cleaning up the spawner's reaper
// closes the connection it holds, as well as terminating its container.
func Test_ReaperCleanupClosesConn(t *testing.T) {
	local, remote := net.Pipe()
	t.Cleanup(func() { require.NoError(t, remote.Close()) })

	held := &reaperConn{}
	held.set(local, "127.0.0.1:8080")

	ctr := &terminateRecorder{terminated: make(chan struct{})}
	reaper := &Reaper{SessionID: testSessionID, container: ctr}
	require.NoError(t, reaper.keep(held))

	spawner := newReaperSpawner()
	spawner.instance = reaper
	require.NoError(t, spawner.cleanup())

	select {
	case <-ctr.terminated:
	default:
		t.Fatal("the container should be terminated")
	}

	errs := make(chan error, 1)
	go func() {
		_, err := remote.Read(make([]byte, 1))
		errs <- err
	}()

	select {
	case err := <-errs:
		require.ErrorIs(t, err, io.EOF, "the connection should be closed")
	case <-time.After(time.Second * 5):
		t.Fatal("the connection should be closed")
	}
}

// Test_ReaperKeep tests that a Reaper keeps the connection a reaper
// acknowledged its filters on, until it's signalled to close it.
func Test_ReaperKeep(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		reaper := &Reaper{SessionID: testSessionID}
		require.Error(t, reaper.keep(&reaperConn{}))
		require.Nil(t, reaper.useTermSignal(), "no connection should be kept")
	})

	t.Run("held", func(t *testing.T) {
		local, remote := net.Pipe()
		t.Cleanup(func() { require.NoError(t, remote.Close()) })

		held := &reaperConn{}
		held.set(local, "127.0.0.1:8080")

		reaper := &Reaper{SessionID: testSessionID}
		require.NoError(t, reaper.keep(held))
		require.Equal(t, "127.0.0.1:8080", reaper.Endpoint, "the connection's endpoint should be the Reaper's")
		conn, _ := held.take()
		require.Nil(t, conn, "the connection should have been taken")

		termSignal := reaper.useTermSignal()
		require.NotNil(t, termSignal, "the connection should be the Reaper's")

		termSignal <- true
		_, err := remote.Read(make([]byte, 1))
		require.ErrorIs(t, err, io.EOF, "the connection should be closed when signalled")
	})
}

// Test_NewReaperRyukPort tests that a reaper created with RYUK_PORT set
// listens on that port, so the session's resources are reaped.
func Test_NewReaperRyukPort(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	t.Setenv("RYUK_PORT", "9000")

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	sessionID := testSessionID + "-ryuk-port"
	t.Cleanup(func() {
		_, err := provider.Client().ContainerRemove(context.Background(), reaperContainerNameFromSessionID(sessionID), client.ContainerRemoveOptions{Force: true})
		require.True(t, isCleanupSafe(err), "remove reaper: %v", err)
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	spawner := newReaperSpawner()
	reaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.host), sessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")

	reaperConnect(t, reaper)
}

// Test_RunReaperReportsAbandonedFailure tests that a run which fails after
// its caller has returned reports the failure, including a failure to remove
// its container, as nobody is waiting for its result.
func Test_RunReaperReportsAbandonedFailure(t *testing.T) {
	provider := &blockingReaperProvider{
		mockReaperProvider: newMockReaperProvider(config.Config{}),
		started:            make(chan context.Context, 1),
		release:            make(chan error, 1),
		ctr:                failingContainer{},
	}

	callerLogger := &recordingLogger{}
	backgroundLogger := &recordingLogger{}
	logger := &detachableLogger{logger: callerLogger, background: backgroundLogger}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errs := make(chan error, 1)
	go func() {
		_, err := newReaperSpawner().runReaper(ctx, provider, ContainerRequest{}, &reaperConn{}, logger, time.Minute)
		errs <- err
	}()

	select {
	case <-provider.started:
	case <-time.After(time.Second * 10):
		t.Fatal("reaper run should have started")
	}

	cancel()

	select {
	case err := <-errs:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second * 5):
		t.Fatal("runReaper should return when its caller's context is done")
	}

	provider.release <- errExpected

	require.Eventually(t, func() bool {
		for _, line := range backgroundLogger.Lines() {
			if strings.Contains(line, "failing-container") &&
				strings.Contains(line, errExpected.Error()) &&
				strings.Contains(line, errTerminate.Error()) {
				return true
			}
		}
		return false
	}, time.Second*5, time.Millisecond*10, "the failure should be reported with its container and cleanup error")
	require.Empty(t, callerLogger.Lines(), "the caller's logger should not be used after it returned")
}

// cancellingLogger is a recordingLogger that calls cancel when it first logs
// a line containing trigger.
type cancellingLogger struct {
	recordingLogger
	trigger string
	once    sync.Once
	cancel  context.CancelFunc
}

func (l *cancellingLogger) Printf(format string, v ...any) {
	l.recordingLogger.Printf(format, v...)
	if strings.Contains(fmt.Sprintf(format, v...), l.trigger) {
		l.once.Do(l.cancel)
	}
}

// Test_NewReaperDetachesCallerLogger tests that a reaper run that outlives
// its caller stops logging to the provider's logger, which may be a test
// logger that can't be used once its test has completed.
func Test_NewReaperDetachesCallerLogger(t *testing.T) {
	reaperDisable(t, false)

	SkipIfProviderIsNotHealthy(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cancels the caller once the run is creating the reaper container.
	logger := &cancellingLogger{trigger: "Creating container", cancel: cancel}
	provider, err := NewDockerProvider(WithLogger(logger))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, provider.Close()) })

	// A session of its own, so the run left in the background doesn't affect
	// other tests. The reaper it starts is removed once the run is complete.
	sessionID := testSessionID + "-detach-logger"
	t.Cleanup(func() {
		_, err := provider.Client().ContainerRemove(context.Background(), reaperContainerNameFromSessionID(sessionID), client.ContainerRemoveOptions{Force: true})
		require.True(t, isCleanupSafe(err), "remove reaper: %v", err)
	})

	background := &recordingLogger{}
	spawner := newReaperSpawner()
	spawner.detachedLogger = background
	_, err = spawner.newReaper(context.WithValue(ctx, core.DockerHostContextKey, provider.host), sessionID, provider)
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorContains(t, err, "run container", "the caller should have been cancelled during the run")

	logged := len(logger.Lines())

	// Wait for the run to complete, whichever logger it uses, so the checks
	// below cover everything it logs. Its last line is the readiness log if
	// it succeeds, or the report of its failure if not.
	completed := func(lines []string) bool {
		for _, line := range lines {
			if strings.Contains(line, "Container is ready") || strings.Contains(line, "failed to run after its caller returned") {
				return true
			}
		}

		return false
	}
	require.Eventually(t, func() bool {
		return completed(logger.Lines()) || completed(background.Lines())
	}, reaperReadyTimeout+reaperStartTimeout+reaperCleanupTimeout+time.Second*10, time.Millisecond*100, "the reaper run should complete")

	require.Len(t, logger.Lines(), logged, "the run should not log to the provider's logger after its caller returned")
	require.True(t, completed(background.Lines()), "the run should log to the background logger after its caller returned")
}

// imageInspectRecorder is an APIClient that records the images inspected,
// reporting each as present.
type imageInspectRecorder struct {
	client.APIClient
	images []string
}

func (c *imageInspectRecorder) ImageInspect(_ context.Context, image string, _ ...client.ImageInspectOption) (client.ImageInspectResult, error) {
	c.images = append(c.images, image)
	return client.ImageInspectResult{}, nil
}

// registrySubstitutor is an ImageSubstitutor that pulls images from registry.
type registrySubstitutor struct {
	registry string
}

func (s registrySubstitutor) Description() string {
	return "registrySubstitutor (prepends " + s.registry + ")"
}

func (s registrySubstitutor) Substitute(image string) (string, error) {
	return path.Join(s.registry, image), nil
}

// Test_PullReaperImage tests that the reaper image is pulled under the name
// CreateContainer would use.
func Test_PullReaperImage(t *testing.T) {
	tests := map[string]struct {
		prefix       string
		substitutors []ImageSubstitutor
		want         string
	}{
		"default": {
			want: config.ReaperDefaultImage,
		},
		"hub-prefix": {
			prefix: "registry.example/mirror",
			want:   path.Join("registry.example/mirror", config.ReaperDefaultImage),
		},
		"request-substitutor": {
			prefix:       "registry.example/mirror",
			substitutors: []ImageSubstitutor{registrySubstitutor{registry: "other.example"}},
			want:         path.Join("other.example", config.ReaperDefaultImage),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cli := &imageInspectRecorder{}
			provider := &DockerProvider{
				DockerProviderOptions: &DockerProviderOptions{GenericProviderOptions: &GenericProviderOptions{Logger: log.TestLogger(t)}},
				client:                cli,
				config:                config.Config{HubImageNamePrefix: tc.prefix},
			}
			req := ContainerRequest{
				Image:             config.ReaperDefaultImage,
				ImageSubstitutors: tc.substitutors,
			}

			require.NoError(t, pullReaperImage(context.Background(), provider, req))
			require.Equal(t, []string{tc.want}, cli.images)
		})
	}
}

// startReaperTestContainer starts an alpine container running cmd and returns
// its ID. If asReaper is true, it has the name and labels of the test
// session's reaper, so lookupContainer finds it.
func startReaperTestContainer(ctx context.Context, t *testing.T, provider *DockerProvider, asReaper bool, cmd ...string) string {
	t.Helper()

	id := createReaperTestContainer(ctx, t, provider, asReaper, cmd...)
	_, err := provider.Client().ContainerStart(ctx, id, client.ContainerStartOptions{})
	require.NoError(t, err)

	return id
}

// createReaperTestContainer creates, but does not start, an alpine container
// running cmd and returns its ID. If asReaper is true, it has the name and
// labels of the test session's reaper, so lookupContainer finds it.
func createReaperTestContainer(ctx context.Context, t *testing.T, provider *DockerProvider, asReaper bool, cmd ...string) string {
	t.Helper()

	opts := client.ContainerCreateOptions{
		Config: &container.Config{
			Image: alpineImage,
			Cmd:   cmd,
		},
	}
	if asReaper {
		labels := core.DefaultLabels(testSessionID)
		labels[core.LabelReaper] = "true"
		labels[core.LabelRyuk] = "true"
		delete(labels, core.LabelReap)

		opts.Config.Labels = labels
		opts.Name = reaperContainerNameFromSessionID(testSessionID)
	}

	return createTestContainer(ctx, t, provider, opts)
}

// startUnreadyReaperContainer starts an alpine container with the name and
// labels of the test session's reaper, so lookupContainer finds it, and
// returns its ID. It exposes the reaper port, published if publish is true,
// but nothing listens on it, so it never becomes ready.
func startUnreadyReaperContainer(ctx context.Context, t *testing.T, provider *DockerProvider, publish bool) string {
	t.Helper()

	labels := core.DefaultLabels(testSessionID)
	labels[core.LabelReaper] = "true"
	labels[core.LabelRyuk] = "true"
	delete(labels, core.LabelReap)

	id := createTestContainer(ctx, t, provider, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:        alpineImage,
			Cmd:          []string{"sleep", "300"},
			Labels:       labels,
			ExposedPorts: network.PortSet{defaultReaperPort: struct{}{}},
		},
		HostConfig: &container.HostConfig{
			PublishAllPorts: publish,
		},
		Name: reaperContainerNameFromSessionID(testSessionID),
	})

	_, err := provider.Client().ContainerStart(ctx, id, client.ContainerStartOptions{})
	require.NoError(t, err)

	return id
}

// createTestContainer creates, but does not start, a container from opts,
// which is removed when the test completes, and returns its ID.
func createTestContainer(ctx context.Context, t *testing.T, provider *DockerProvider, opts client.ContainerCreateOptions) string {
	t.Helper()

	require.NoError(t, provider.PullImage(ctx, alpineImage))

	cli := provider.Client()
	created, err := cli.ContainerCreate(ctx, opts)
	require.NoError(t, err)
	t.Cleanup(func() {
		if _, err := cli.ContainerRemove(context.Background(), created.ID, client.ContainerRemoveOptions{Force: true}); err != nil && !errdefs.IsNotFound(err) {
			require.NoError(t, err)
		}
	})

	return created.ID
}

// requireContainerStopped waits for the container to stop running.
func requireContainerStopped(ctx context.Context, t *testing.T, cli client.APIClient, id string) {
	t.Helper()

	require.Eventually(t, func() bool {
		inspect, err := cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
		return err == nil && !inspect.Container.State.Running
	}, time.Second*10, time.Millisecond*100, "container should have stopped")
}

func TestReaper_reuseItFromOtherTestProgramUsingDocker(t *testing.T) {
	reaperDisable(t, false)

	// Explicitly set the reaper instance to nil to simulate another test
	// program in the same session accessing the same reaper.
	spawner.instance = nil

	SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()
	// As other integration tests run with the (shared) Reaper as well,
	// re-use the instance to not interrupt other tests.
	if spawner.instance != nil {
		t.Cleanup(func() {
			require.NoError(t, spawner.cleanup())
		})
	}

	provider, err := ProviderDefault.GetProvider()
	require.NoError(t, err)

	reaper, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.(*DockerProvider).host), testSessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "creating the Reaper should not error")

	// Explicitly reset the reaper instance to nil to simulate another test
	// program in the same session accessing the same reaper.
	spawner.instance = nil

	reaperReused, err := spawner.reaper(context.WithValue(ctx, core.DockerHostContextKey, provider.(*DockerProvider).host), testSessionID, provider)
	cleanupReaper(t, reaper, spawner)
	require.NoError(t, err, "reusing the Reaper should not error")

	// Ensure that the internal state of both reaper instances is the same.
	require.Equal(t, reaper.SessionID, reaperReused.SessionID, "expecting the same SessionID")
	require.Equal(t, reaper.Endpoint, reaperReused.Endpoint, "expecting the same reaper endpoint")
	require.Equal(t, reaper.Provider, reaperReused.Provider, "expecting the same container provider")
	require.Equal(t, reaper.container.GetContainerID(), reaperReused.container.GetContainerID(), "expecting the same container ID")
	require.Equal(t, reaper.container.SessionID(), reaperReused.container.SessionID(), "expecting the same session ID")

	termSignal, err := reaper.Connect()
	cleanupTermSignal(t, termSignal)
	require.NoError(t, err, "connecting to Reaper should be successful")
}

// TestReaper_ReuseRunning tests whether reusing the reaper if using
// testcontainers from concurrently multiple packages works as expected. In this
// case, global locks are without any effect as Go tests different packages
// isolated. Therefore, this test does not use the same logic with locks on
// purpose. We expect reaper creation to still succeed in case a reaper is
// already running for the same session id by returning its container instance
// instead.
func TestReaper_ReuseRunning(t *testing.T) {
	reaperDisable(t, false)

	const concurrency = 64

	timeout, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	sessionID := SessionID()

	dockerProvider, err := NewDockerProvider()
	require.NoError(t, err, "new docker provider should not fail")

	obtainedReaperContainerIDs := make([]string, concurrency)
	var wg sync.WaitGroup
	for i := range concurrency {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			spawner := newReaperSpawner()
			reaper, err := spawner.reaper(timeout, sessionID, dockerProvider)
			cleanupReaper(t, reaper, spawner)
			require.NoError(t, err)

			reaperConnect(t, reaper)

			obtainedReaperContainerIDs[i] = reaper.container.GetContainerID()
		}(i)
	}
	wg.Wait()

	// Assure that all calls returned the same container.
	firstContainerID := obtainedReaperContainerIDs[0]
	for i, containerID := range obtainedReaperContainerIDs {
		require.Equal(t, firstContainerID, containerID, "call %d should have returned same container id", i)
	}
}

// reaperConnect requires that connecting to reaper succeeds.
func reaperConnect(t *testing.T, reaper *Reaper) {
	t.Helper()

	termSignal, err := reaper.connect(context.Background())
	require.NoError(t, err, "Reaper handshake should be successful")
	termSignal <- true
}

func TestSpawnerBackoff(t *testing.T) {
	b := spawner.backoff()
	for range 100 {
		require.LessOrEqual(t, b.NextBackOff(), time.Millisecond*250, "backoff should not exceed max interval")
	}
}

type timeoutErr struct{}

func (timeoutErr) Error() string {
	return "timeout"
}

func (timeoutErr) Timeout() bool {
	return true
}

func TestSpawnerRetryError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		err := spawner.retryError(nil)
		require.NoError(t, err, "should return nil")
	})

	tests := []struct {
		name      string
		err       error
		permanent bool
	}{
		{
			name:      "docker conflict error",
			err:       errors.New("Conflict. The container name \"foo\" is already in use by container \"01234\"."),
			permanent: false,
		},
		{
			name:      "podman conflict error",
			err:       errors.New("creating container storage: the container name \"foo\" is already in use by 01234."),
			permanent: false,
		},
		{
			name:      "errdefs.ErrNotFound",
			err:       fmt.Errorf("foo: %w", errdefs.ErrNotFound),
			permanent: false,
		},
		{
			name:      "errdefs.Conflict",
			err:       fmt.Errorf("foo: %w", errdefs.ErrConflict),
			permanent: true,
		},
		{
			name:      "syscall.ECONNREFUSED",
			err:       fmt.Errorf("foo: %w", syscall.ECONNREFUSED),
			permanent: false,
		},
		{
			name:      "syscall.ECONNRESET",
			err:       fmt.Errorf("foo: %w", syscall.ECONNRESET),
			permanent: false,
		},
		{
			name:      "syscall.ECONNABORTED",
			err:       fmt.Errorf("foo: %w", syscall.ECONNABORTED),
			permanent: false,
		},
		{
			name:      "syscall.ETIMEDOUT",
			err:       fmt.Errorf("foo: %w", syscall.ETIMEDOUT),
			permanent: false,
		},
		{
			name:      "os.ErrDeadlineExceeded",
			err:       fmt.Errorf("foo: %w", os.ErrDeadlineExceeded),
			permanent: false,
		},
		{
			name:      "context.DeadlineExceeded",
			err:       fmt.Errorf("foo: %w", context.DeadlineExceeded),
			permanent: false,
		},
		{
			name:      "timeout error",
			err:       fmt.Errorf("foo: %w", timeoutErr{}),
			permanent: false,
		},
		{
			name:      "context.Canceled",
			err:       fmt.Errorf("foo: %w", context.Canceled),
			permanent: false,
		},
		{
			name:      "wait failure on a stopped container",
			err:       fmt.Errorf("foo: %w", errors.Join(errors.New("container exited with code 0"), errdefs.ErrNotFound.WithMessage("container state: exited"))),
			permanent: false,
		},
		{
			name:      "random error",
			err:       errors.New("some random error"),
			permanent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := spawner.retryError(tt.err)
			require.Error(t, gotErr, "should not return nil")

			permanentError := &backoff.PermanentError{}
			isPermanent := errors.As(gotErr, &permanentError)
			if tt.permanent {
				require.True(t, isPermanent, "the error should be a PermanentError")
			} else {
				require.False(t, isPermanent, "the error should not be a PermanentError")
			}
		})
	}
}

// cleanupReaper schedules reaper for cleanup if it's not nil.
func cleanupReaper(t *testing.T, reaper *Reaper, spawner *reaperSpawner) {
	t.Helper()

	if reaper == nil {
		return
	}

	t.Cleanup(func() {
		reaper.close()
		require.NoError(t, spawner.cleanup())
	})
}

// cleanupTermSignal ensures that termSignal
func cleanupTermSignal(t *testing.T, termSignal chan bool) {
	t.Helper()

	t.Cleanup(func() {
		if termSignal != nil {
			termSignal <- true
		}
	})
}
