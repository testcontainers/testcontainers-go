package testcontainers

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/internal/testcontainerssession"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testSessionID the tests need to create a reaper in a different session, so that it does not interfere with other tests
const testSessionID = "this-is-a-different-session-id"

type mockReaperProvider struct {
	req               ContainerRequest
	hostConfig        *container.HostConfig
	enpointSettings   map[string]*network.EndpointSettings
	config            TestcontainersConfig
	initialReaper     *Reaper
	initialReaperOnce sync.Once
	t                 *testing.T
}

func newMockReaperProvider(t *testing.T) *mockReaperProvider {
	m := &mockReaperProvider{
		config: TestcontainersConfig{
			Config: config.Config{},
		},
		t:             t,
		initialReaper: reaperInstance,
		//nolint:govet
		initialReaperOnce: reaperOnce,
	}

	// explicitly reset the reaperInstance to nil to start from a fresh state
	reaperInstance = nil
	reaperOnce = sync.Once{}

	return m
}

var errExpected = errors.New("expected")

func (m *mockReaperProvider) RestoreReaperState() {
	m.t.Cleanup(func() {
		reaperInstance = m.initialReaper
		//nolint:govet
		reaperOnce = m.initialReaperOnce
	})
}

func (m *mockReaperProvider) RunContainer(ctx context.Context, req ContainerRequest) (Container, error) {
	m.req = req

	m.hostConfig = &container.HostConfig{}
	m.enpointSettings = map[string]*network.EndpointSettings{}

	if req.HostConfigModifier == nil {
		req.HostConfigModifier = defaultHostConfigModifier(req)
	}
	req.HostConfigModifier(m.hostConfig)

	if req.EnpointSettingsModifier != nil {
		req.EnpointSettingsModifier(m.enpointSettings)
	}

	// we're only interested in the request, so instead of mocking the Docker client
	// we'll error out here
	return nil, errExpected
}

func (m *mockReaperProvider) Config() TestcontainersConfig {
	return m.config
}

// createContainerRequest creates the expected request and allows for customization
func createContainerRequest(customize func(ContainerRequest) ContainerRequest) ContainerRequest {
	req := ContainerRequest{
		Image:        config.ReaperDefaultImage,
		ExposedPorts: []string{"8080/tcp"},
		Labels:       testcontainersdocker.DefaultLabels(testSessionID),
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.Binds = []string{testcontainersdocker.ExtractDockerSocket(context.Background()) + ":/var/run/docker.sock"}
		},
		WaitingFor: wait.ForListeningPort(nat.Port("8080/tcp")),
		Env: map[string]string{
			"RYUK_CONNECTION_TIMEOUT":   "1m0s",
			"RYUK_RECONNECTION_TIMEOUT": "10s",
		},
	}

	req.Labels[testcontainersdocker.LabelReaper] = "true"
	req.Labels[testcontainersdocker.LabelRyuk] = "true"

	if customize == nil {
		return req
	}

	return customize(req)
}

func TestContainerStartsWithoutTheReaper(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	tcConfig := config.Read()
	if !tcConfig.RyukDisabled {
		t.Skip("Ryuk is enabled, skipping test")
	}

	ctx := context.Background()

	container, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
		},
		Started: true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	sessionID := testcontainerssession.SessionID()

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
	tcConfig := config.Read()
	if tcConfig.RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	ctx := context.Background()

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	terminateContainerOnEnd(t, ctx, c)

	sessionID := testcontainerssession.SessionID()

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
	tcConfig := config.Read()
	if tcConfig.RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

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

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, nginxA)

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
	tcConfig := config.Read()
	if tcConfig.RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

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
	tcConfig := config.Read()
	if !tcConfig.RyukDisabled {
		t.Skip("Ryuk is enabled, skipping test")
	}

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

func Test_NewReaper(t *testing.T) {
	type cases struct {
		name   string
		req    ContainerRequest
		config TestcontainersConfig
		ctx    context.Context
		env    map[string]string
	}

	tests := []cases{
		{
			name: "non-privileged",
			req:  createContainerRequest(nil),
			config: TestcontainersConfig{Config: config.Config{
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Second,
			}},
		},
		{
			name: "privileged",
			req: createContainerRequest(func(req ContainerRequest) ContainerRequest {
				req.Privileged = true
				return req
			}),
			config: TestcontainersConfig{Config: config.Config{
				RyukPrivileged:          true,
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Second,
			}},
		},
		{
			name: "configured non-default timeouts",
			req: createContainerRequest(func(req ContainerRequest) ContainerRequest {
				req.Env = map[string]string{
					"RYUK_CONNECTION_TIMEOUT":   "1m0s",
					"RYUK_RECONNECTION_TIMEOUT": "10m0s",
				}
				return req
			}),
			config: TestcontainersConfig{Config: config.Config{
				RyukPrivileged:          true,
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Minute,
			}},
		},
		{
			name: "docker-host in context",
			req: createContainerRequest(func(req ContainerRequest) ContainerRequest {
				req.HostConfigModifier = func(hostConfig *container.HostConfig) {
					hostConfig.Binds = []string{testcontainersdocker.ExtractDockerSocket(context.Background()) + ":/var/run/docker.sock"}
				}
				return req
			}),
			config: TestcontainersConfig{Config: config.Config{
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Second,
			}},
			ctx: context.WithValue(context.TODO(), testcontainersdocker.DockerHostContextKey, testcontainersdocker.DockerSocketPathWithSchema),
		},
		{
			name: "Reaper including custom Hub prefix",
			req: createContainerRequest(func(req ContainerRequest) ContainerRequest {
				req.Image = config.ReaperDefaultImage
				req.Privileged = true
				return req
			}),
			config: TestcontainersConfig{Config: config.Config{
				HubImageNamePrefix:      "registry.mycompany.com/mirror",
				RyukPrivileged:          true,
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Second,
			}},
		},
		{
			name: "Reaper including custom Hub prefix as env var",
			req: createContainerRequest(func(req ContainerRequest) ContainerRequest {
				req.Image = config.ReaperDefaultImage
				req.Privileged = true
				return req
			}),
			config: TestcontainersConfig{Config: config.Config{
				RyukPrivileged:          true,
				RyukConnectionTimeout:   time.Minute,
				RyukReconnectionTimeout: 10 * time.Second,
			}},
			env: map[string]string{
				"TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX": "registry.mycompany.com/mirror",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.env != nil {
				config.Reset() // reset the config using the internal method to avoid the sync.Once
				for k, v := range test.env {
					t.Setenv(k, v)
				}
			}

			if prefix := os.Getenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX"); prefix != "" {
				test.config.Config.HubImageNamePrefix = prefix
			}

			provider := newMockReaperProvider(t)
			provider.config = test.config
			t.Cleanup(provider.RestoreReaperState)

			if test.ctx == nil {
				test.ctx = context.TODO()
			}

			_, err := reuseOrCreateReaper(test.ctx, testSessionID, provider)
			// we should have errored out see mockReaperProvider.RunContainer
			assert.EqualError(t, err, "expected")

			assert.Equal(t, test.req.Image, provider.req.Image, "expected image doesn't match the submitted request")
			assert.Equal(t, test.req.ExposedPorts, provider.req.ExposedPorts, "expected exposed ports don't match the submitted request")
			assert.Equal(t, test.req.Labels, provider.req.Labels, "expected labels don't match the submitted request")
			assert.Equal(t, test.req.Mounts, provider.req.Mounts, "expected mounts don't match the submitted request")
			assert.Equal(t, test.req.WaitingFor, provider.req.WaitingFor, "expected waitingFor don't match the submitted request")
			assert.Equal(t, test.req.Env, provider.req.Env, "expected env doesn't match the submitted request")

			// checks for reaper's preCreationCallback fields
			assert.Equal(t, container.NetworkMode(Bridge), provider.hostConfig.NetworkMode, "expected networkMode doesn't match the submitted request")
			assert.Equal(t, true, provider.hostConfig.AutoRemove, "expected networkMode doesn't match the submitted request")
		})
	}
}

func Test_ReaperReusedIfHealthy(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	tcConfig := config.Read()
	if tcConfig.RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	testProvider := newMockReaperProvider(t)
	t.Cleanup(testProvider.RestoreReaperState)

	SkipIfProviderIsNotHealthy(&testing.T{})

	ctx := context.Background()
	// As other integration tests run with the (shared) Reaper as well, re-use the instance to not interrupt other tests
	wasReaperRunning := reaperInstance != nil

	provider, _ := ProviderDocker.GetProvider()
	reaper, err := reuseOrCreateReaper(context.WithValue(ctx, testcontainersdocker.DockerHostContextKey, provider.(*DockerProvider).host), testSessionID, provider)
	assert.NoError(t, err, "creating the Reaper should not error")

	reaperReused, err := reuseOrCreateReaper(context.WithValue(ctx, testcontainersdocker.DockerHostContextKey, provider.(*DockerProvider).host), testSessionID, provider)
	assert.NoError(t, err, "reusing the Reaper should not error")
	// assert that the internal state of both reaper instances is the same
	assert.Equal(t, reaper.SessionID, reaperReused.SessionID, "expecting the same SessionID")
	assert.Equal(t, reaper.Endpoint, reaperReused.Endpoint, "expecting the same reaper endpoint")
	assert.Equal(t, reaper.Provider, reaperReused.Provider, "expecting the same container provider")
	assert.Equal(t, reaper.container.GetContainerID(), reaperReused.container.GetContainerID(), "expecting the same container ID")
	assert.Equal(t, reaper.container.SessionID(), reaperReused.container.SessionID(), "expecting the same session ID")

	terminate, err := reaper.Connect()
	defer func(term chan bool) {
		term <- true
	}(terminate)
	assert.NoError(t, err, "connecting to Reaper should be successful")

	if !wasReaperRunning {
		terminateContainerOnEnd(t, ctx, reaper.container)
	}
}

func TestReaper_reuseItFromOtherTestProgramUsingDocker(t *testing.T) {
	config.Reset() // reset the config using the internal method to avoid the sync.Once
	tcConfig := config.Read()
	if tcConfig.RyukDisabled {
		t.Skip("Ryuk is disabled, skipping test")
	}

	mockProvider := &mockReaperProvider{
		initialReaper: reaperInstance,
		//nolint:govet
		initialReaperOnce: reaperOnce,
		t:                 t,
	}
	t.Cleanup(mockProvider.RestoreReaperState)

	// explicitly set the reaperInstance to nil to simulate another test program in the same session accessing the same reaper
	reaperInstance = nil
	reaperOnce = sync.Once{}

	SkipIfProviderIsNotHealthy(&testing.T{})

	ctx := context.Background()
	// As other integration tests run with the (shared) Reaper as well, re-use the instance to not interrupt other tests
	wasReaperRunning := reaperInstance != nil

	provider, _ := ProviderDocker.GetProvider()
	reaper, err := reuseOrCreateReaper(context.WithValue(ctx, testcontainersdocker.DockerHostContextKey, provider.(*DockerProvider).host), testSessionID, provider)
	assert.NoError(t, err, "creating the Reaper should not error")

	// explicitly reset the reaperInstance to nil to simulate another test program in the same session accessing the same reaper
	reaperInstance = nil
	reaperOnce = sync.Once{}

	reaperReused, err := reuseOrCreateReaper(context.WithValue(ctx, testcontainersdocker.DockerHostContextKey, provider.(*DockerProvider).host), testSessionID, provider)
	assert.NoError(t, err, "reusing the Reaper should not error")
	// assert that the internal state of both reaper instances is the same
	assert.Equal(t, reaper.SessionID, reaperReused.SessionID, "expecting the same SessionID")
	assert.Equal(t, reaper.Endpoint, reaperReused.Endpoint, "expecting the same reaper endpoint")
	assert.Equal(t, reaper.Provider, reaperReused.Provider, "expecting the same container provider")
	assert.Equal(t, reaper.container.GetContainerID(), reaperReused.container.GetContainerID(), "expecting the same container ID")
	assert.Equal(t, reaper.container.SessionID(), reaperReused.container.SessionID(), "expecting the same session ID")

	terminate, err := reaper.Connect()
	defer func(term chan bool) {
		term <- true
	}(terminate)
	assert.NoError(t, err, "connecting to Reaper should be successful")

	if !wasReaperRunning {
		terminateContainerOnEnd(t, ctx, reaper.container)
	}
}

// TestReaper_ReuseRunning tests whether reusing the reaper if using
// testcontainers from concurrently multiple packages works as expected. In this
// case, global locks are without any effect as Go tests different packages
// isolated. Therefore, this test does not use the same logic with locks on
// purpose. We expect reaper creation to still succeed in case a reaper is
// already running for the same session id by returning its container instance
// instead.
func TestReaper_ReuseRunning(t *testing.T) {
	const concurrency = 64

	timeout, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sessionID := SessionID()

	dockerProvider, err := NewDockerProvider()
	require.NoError(t, err, "new docker provider should not fail")

	obtainedReaperContainerIDs := make([]string, concurrency)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			reaperContainer, err := lookUpReaperContainer(timeout, sessionID)
			if err == nil && reaperContainer != nil {
				// Found.
				obtainedReaperContainerIDs[i] = reaperContainer.GetContainerID()
				return
			}
			// Not found -> create.
			createdReaper, err := newReaper(timeout, sessionID, dockerProvider)
			require.NoError(t, err, "new reaper should not fail")
			obtainedReaperContainerIDs[i] = createdReaper.container.GetContainerID()
		}()
	}
	wg.Wait()

	// Assure that all calls returned the same container.
	firstContainerID := obtainedReaperContainerIDs[0]
	for i, containerID := range obtainedReaperContainerIDs {
		assert.Equal(t, firstContainerID, containerID, "call %d should have returned same container id", i)
	}
}
