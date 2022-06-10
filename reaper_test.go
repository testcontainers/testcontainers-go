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
		ExposedPorts: []string{"8080/tcp"},
		Labels: map[string]string{
			TestcontainerLabel:         "true",
			TestcontainerLabelIsReaper: "true",
		},
		SkipReaper:  true,
		Mounts:      Mounts(BindMount("/var/run/docker.sock", "/var/run/docker.sock")),
		AutoRemove:  true,
		WaitingFor:  wait.ForListeningPort(nat.Port("8080/tcp")),
		NetworkMode: "bridge",
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
		config TestContainersConfig
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// make sure we re-initialize the singleton
			reaper = nil
			provider := &mockReaperProvider{
				config: test.config,
			}

			_, err := NewReaper(context.TODO(), "sessionId", provider, "reaperImage")
			// we should have errored out see mockReaperProvider.RunContainer
			assert.EqualError(t, err, "expected")

			assert.Equal(t, test.req, provider.req, "expected ContainerRequest doesn't match the submitted request")
		})
	}
}
