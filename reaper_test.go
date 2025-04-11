package testcontainers

import (
	"context"
	"errors"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
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
		WaitingFor: wait.ForListeningPort(nat.Port("8080/tcp")),
		Env: map[string]string{
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

	ctr, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
		},
		Started: true,
	})
	CleanupContainer(t, ctr)
	require.NoError(t, err)
}

// testReaperRunning validates that a reaper is running.
func testReaperRunning(t *testing.T) {
	t.Helper()

	ctx := context.Background()
	sessionID := core.SessionID()
	reaperContainer, err := spawner.lookupContainer(ctx, sessionID)
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

	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
		},
		Started: true,
	})
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
	require.Equal(t, "exited", state.Status)
}

// testContainerTerminate tests terminating a container.
func testContainerTerminate(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	nginxA, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
		},
		Started: true,
	})
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
	spawner := &reaperSpawner{}
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

	provider, err := ProviderDocker.GetProvider()
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

	provider, err := ProviderDocker.GetProvider()
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
		state, err := reaper.container.State(ctx)
		if err != nil {
			if errdefs.IsNotFound(err) {
				break
			}
			require.NoError(t, err)
		}

		if !state.Running {
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

	provider, err := ProviderDocker.GetProvider()
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
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			spawner := &reaperSpawner{}
			reaper, err := spawner.reaper(timeout, sessionID, dockerProvider)
			cleanupReaper(t, reaper, spawner)
			require.NoError(t, err)

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

func TestSpawnerBackoff(t *testing.T) {
	b := spawner.backoff()
	for i := 0; i < 100; i++ {
		require.LessOrEqual(t, b.NextBackOff(), time.Millisecond*250, "backoff should not exceed max interval")
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
