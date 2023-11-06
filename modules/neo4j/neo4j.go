package neo4j

import (
	"context"
	"fmt"
	"net/http"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// defaultImage {
	defaultImageName = "neo4j"
	defaultTag       = "4.4"
	// }
)

const (
	// containerPorts {
	defaultBoltPort  = "7687"
	defaultHttpPort  = "7474"
	defaultHttpsPort = "7473"
	// }
)

// Neo4jContainer represents the Neo4j container type used in the module
type Neo4jContainer struct {
	testcontainers.Container
}

// BoltUrl returns the bolt url for the Neo4j container, using the bolt port, in the format of neo4j://host:port
func (c Neo4jContainer) BoltUrl(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	containerPort, err := nat.NewPort("tcp", defaultBoltPort)
	if err != nil {
		return "", err
	}
	mappedPort, err := c.MappedPort(ctx, containerPort)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("neo4j://%s:%d", host, mappedPort.Int()), nil
}

// RunContainer creates an instance of the Neo4j container type
func RunContainer(ctx context.Context, options ...testcontainers.ContainerCustomizer) (*Neo4jContainer, error) {
	httpPort, _ := nat.NewPort("tcp", defaultHttpPort)
	request := testcontainers.ContainerRequest{
		Image: fmt.Sprintf("docker.io/%s:%s", defaultImageName, defaultTag),
		Env: map[string]string{
			"NEO4J_AUTH": "none",
		},
		ExposedPorts: []string{
			fmt.Sprintf("%s/tcp", defaultBoltPort),
			fmt.Sprintf("%s/tcp", defaultHttpPort),
			fmt.Sprintf("%s/tcp", defaultHttpsPort),
		},
		WaitingFor: &wait.MultiStrategy{
			Strategies: []wait.Strategy{
				wait.NewLogStrategy("Bolt enabled on"),
				&wait.HTTPStrategy{
					Port:              httpPort,
					StatusCodeMatcher: isHttpOk(),
				},
			},
		},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Logger:           testcontainers.Logger,
		Started:          true,
	}

	if len(options) == 0 {
		options = append(options, WithoutAuthentication())
	}

	for _, option := range options {
		option.Customize(&genericContainerReq)
	}

	err := validate(&genericContainerReq)
	if err != nil {
		return nil, err
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &Neo4jContainer{Container: container}, nil
}

func isHttpOk() func(status int) bool {
	return func(status int) bool {
		return status == http.StatusOK
	}
}
