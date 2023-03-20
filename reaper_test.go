package testcontainers

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/internal"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/wait"
)

type mockReaperProvider struct {
	req             ContainerRequest
	hostConfig      *container.HostConfig
	enpointSettings map[string]*network.EndpointSettings
	config          TestcontainersConfig
}

var errExpected = errors.New("expected")

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
		Image:        "reaperImage",
		ReaperImage:  "reaperImage",
		ExposedPorts: []string{"8080/tcp"},
		Labels: map[string]string{
			TestcontainerLabel:                "true",
			TestcontainerLabelIsReaper:        "true",
			testcontainersdocker.LabelReaper:  "true",
			testcontainersdocker.LabelLang:    "go",
			testcontainersdocker.LabelVersion: internal.Version,
		},
		Mounts:     Mounts(BindMount("/var/run/docker.sock", "/var/run/docker.sock")),
		WaitingFor: wait.ForListeningPort(nat.Port("8080/tcp")),
		ReaperOptions: []ContainerOption{
			WithImageName("reaperImage"),
		},
	}
	if customize == nil {
		return req
	}

	return customize(req)
}

func Test_NewReaper(t *testing.T) {
	type cases struct {
		name   string
		req    ContainerRequest
		config TestcontainersConfig
		ctx    context.Context
	}

	tests := []cases{
		{
			name:   "non-privileged",
			req:    createContainerRequest(nil),
			config: TestcontainersConfig{},
		},
		{
			name: "privileged",
			req: createContainerRequest(func(req ContainerRequest) ContainerRequest {
				req.Privileged = true
				return req
			}),
			config: TestcontainersConfig{
				RyukPrivileged: true,
			},
		},
		{
			name: "docker-host in context",
			req: createContainerRequest(func(req ContainerRequest) ContainerRequest {
				req.Mounts = Mounts(BindMount("/value/in/context.sock", "/var/run/docker.sock"))
				return req
			}),
			config: TestcontainersConfig{},
			ctx:    context.WithValue(context.TODO(), testcontainersdocker.DockerHostContextKey, "unix:///value/in/context.sock"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &mockReaperProvider{
				config: test.config,
			}

			if test.ctx == nil {
				test.ctx = context.TODO()
			}

			_, err := newReaper(test.ctx, "sessionId", provider, test.req.ReaperOptions...)
			// we should have errored out see mockReaperProvider.RunContainer
			assert.EqualError(t, err, "expected")

			assert.Equal(t, test.req.Image, provider.req.Image, "expected image doesn't match the submitted request")
			assert.Equal(t, test.req.ExposedPorts, provider.req.ExposedPorts, "expected exposed ports don't match the submitted request")
			assert.Equal(t, test.req.Labels, provider.req.Labels, "expected labels don't match the submitted request")
			assert.Equal(t, test.req.Mounts, provider.req.Mounts, "expected mounts don't match the submitted request")
			assert.Equal(t, test.req.WaitingFor, provider.req.WaitingFor, "expected waitingFor don't match the submitted request")

			// checks for reaper's preCreationCallback fields
			assert.Equal(t, container.NetworkMode(Bridge), provider.hostConfig.NetworkMode, "expected networkMode doesn't match the submitted request")
			assert.Equal(t, true, provider.hostConfig.AutoRemove, "expected networkMode doesn't match the submitted request")
		})
	}
}

func Test_ReaperForNetwork(t *testing.T) {
	ctx := context.Background()

	networkName := "test-network-with-custom-reaper"

	req := GenericNetworkRequest{
		NetworkRequest: NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
			ReaperOptions: []ContainerOption{
				WithImageName("reaperImage"),
			},
		},
	}

	provider := &mockReaperProvider{
		config: TestcontainersConfig{},
	}

	_, err := newReaper(ctx, "sessionId", provider, req.ReaperOptions...)
	assert.EqualError(t, err, "expected")

	assert.Equal(t, "reaperImage", provider.req.Image)
	assert.Equal(t, "reaperImage", provider.req.ReaperImage)
}

func Test_ReaperReusedIfHealthy(t *testing.T) {
	SkipIfProviderIsNotHealthy(&testing.T{})

	ctx := context.Background()
	// As other integration tests run with the (shared) Reaper as well, re-use the instance to not interrupt other tests
	wasReaperRunning := reaperInstance != nil

	provider, _ := ProviderDocker.GetProvider()
	reaper, err := reuseOrCreateReaper(context.WithValue(ctx, testcontainersdocker.DockerHostContextKey, provider.(*DockerProvider).host), "sessionId", provider)
	assert.NoError(t, err, "creating the Reaper should not error")

	reaperReused, _ := reuseOrCreateReaper(context.WithValue(ctx, testcontainersdocker.DockerHostContextKey, provider.(*DockerProvider).host), "sessionId", provider)
	assert.Same(t, reaper, reaperReused, "expecting the same reaper instance is returned if running and healthy")

	terminate, err := reaper.Connect()
	defer func(term chan bool) {
		term <- true
	}(terminate)
	assert.NoError(t, err, "connecting to Reaper should be successful")

	if !wasReaperRunning {
		terminateContainerOnEnd(t, ctx, reaper.container)
	}
}
