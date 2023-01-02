package testcontainers

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/wait"
)

type mockReaperProvider struct {
	req    ContainerRequest
	config TestContainersConfig
}

var errExpected = errors.New("expected")

func (m *mockReaperProvider) RunContainer(ctx context.Context, req ContainerRequest) (Container, error) {
	m.req = req

	// we're only interested in the request, so instead of mocking the Docker client
	// we'll error out here
	return nil, errExpected
}

func (m *mockReaperProvider) Config() TestContainersConfig {
	return m.config
}

// createContainerRequest creates the expected request and allows for customization
func createContainerRequest(customize func(ContainerRequest) ContainerRequest) ContainerRequest {
	req := ContainerRequest{
		Image:        "reaperImage",
		ReaperImage:  "reaperImage",
		ExposedPorts: []string{"8080/tcp"},
		Labels: map[string]string{
			TestcontainerLabel:          "true",
			TestcontainerLabelIsReaper:  "true",
			TestcontainerLabelSessionID: "sessionId",
		},
		SkipReaper:  true,
		Mounts:      Mounts(BindMount("/var/run/docker.sock", "/var/run/docker.sock")),
		AutoRemove:  true,
		WaitingFor:  wait.ForListeningPort(nat.Port("8080/tcp")),
		NetworkMode: "bridge",
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
	defer func() { reaper = nil }()

	type cases struct {
		name   string
		req    ContainerRequest
		config TestContainersConfig
		ctx    context.Context
	}

	tests := []cases{
		{
			name:   "non-privileged",
			req:    createContainerRequest(nil),
			config: TestContainersConfig{},
		},
		{
			name: "privileged",
			req: createContainerRequest(func(req ContainerRequest) ContainerRequest {
				req.Privileged = true
				return req
			}),
			config: TestContainersConfig{
				RyukPrivileged: true,
			},
		},
		{
			name: "docker-host in context",
			req: createContainerRequest(func(req ContainerRequest) ContainerRequest {
				req.Mounts = Mounts(BindMount("/value/in/context.sock", "/var/run/docker.sock"))
				return req
			}),
			config: TestContainersConfig{},
			ctx:    context.WithValue(context.TODO(), dockerHostContextKey, "unix:///value/in/context.sock"),
		},
		{
			name: "with registry credentials",
			req: createContainerRequest(func(req ContainerRequest) ContainerRequest {
				creds := "registry-creds"
				req.RegistryCred = creds
				req.ReaperOptions = append(req.ReaperOptions, WithRegistryCredentials(creds))
				return req
			}),
			config: TestContainersConfig{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// make sure we re-initialize the singleton
			reaper = nil
			provider := &mockReaperProvider{
				config: test.config,
			}

			if test.ctx == nil {
				test.ctx = context.TODO()
			}

			_, err := newReaper(test.ctx, "sessionId", provider, test.req.ReaperOptions...)
			// we should have errored out see mockReaperProvider.RunContainer
			assert.EqualError(t, err, "expected")

			assert.Equal(t, test.req, provider.req, "expected ContainerRequest doesn't match the submitted request")
		})
	}
}

func Test_ExtractDockerHost(t *testing.T) {
	defer func() { reaper = nil }()

	t.Run("Docker Host as environment variable", func(t *testing.T) {
		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/path/to/docker.sock")
		host := extractDockerHost(context.Background())

		assert.Equal(t, "/path/to/docker.sock", host)
	})

	t.Run("Default Docker Host", func(t *testing.T) {
		host := extractDockerHost(context.Background())

		assert.Equal(t, "/var/run/docker.sock", host)
	})

	t.Run("Malformed Docker Host is passed in context", func(t *testing.T) {
		ctx := context.Background()

		host := extractDockerHost(context.WithValue(ctx, dockerHostContextKey, "path-to-docker-sock"))

		assert.Equal(t, "/var/run/docker.sock", host)
	})

	t.Run("Malformed Schema Docker Host is passed in context", func(t *testing.T) {
		ctx := context.Background()

		host := extractDockerHost(context.WithValue(ctx, dockerHostContextKey, "http://path to docker sock"))

		assert.Equal(t, "/var/run/docker.sock", host)
	})

	t.Run("Unix Docker Host is passed in context", func(t *testing.T) {
		ctx := context.Background()

		host := extractDockerHost(context.WithValue(ctx, dockerHostContextKey, "unix:///this/is/a/sample.sock"))

		assert.Equal(t, "/this/is/a/sample.sock", host)
	})
}

func Test_ReaperForNetwork(t *testing.T) {
	defer func() { reaper = nil }()

	ctx := context.Background()

	networkName := "test-network-with-custom-reaper"

	req := GenericNetworkRequest{
		NetworkRequest: NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
			ReaperOptions: []ContainerOption{
				WithRegistryCredentials("credentials"),
				WithImageName("reaperImage"),
			},
		},
	}

	provider := &mockReaperProvider{
		config: TestContainersConfig{},
	}

	_, err := newReaper(ctx, "sessionId", provider, req.ReaperOptions...)
	assert.EqualError(t, err, "expected")

	assert.Equal(t, "credentials", provider.req.RegistryCred)
	assert.Equal(t, "reaperImage", provider.req.Image)
	assert.Equal(t, "reaperImage", provider.req.ReaperImage)
}
