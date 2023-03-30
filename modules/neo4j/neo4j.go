package neo4j

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
)

const defaultImageName = "neo4j"
const defaultTag = "4.4"
const defaultBoltPort = "7687"
const defaultHttpPort = "7474"
const defaultHttpsPort = "7473"

// Neo4jContainer represents the Neo4j container type used in the module
type Neo4jContainer struct {
	testcontainers.Container
}

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

// StartContainer creates an instance of the Neo4j container type
func StartContainer(ctx context.Context, options ...Option) (*Neo4jContainer, error) {
	settings := &config{
		imageCoordinates: fmt.Sprintf("docker.io/%s:%s", defaultImageName, defaultTag),
		adminPassword:    "password",
		neo4jSettings:    map[string]string{},
		logger:           testcontainers.Logger,
	}
	for _, option := range options {
		option(settings)
	}

	if err := settings.validate(); err != nil {
		return nil, err
	}

	httpPort, _ := nat.NewPort("tcp", defaultHttpPort)
	request := testcontainers.ContainerRequest{
		Image: settings.imageCoordinates,
		Env:   settings.exportEnv(),
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
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
		Logger:           settings.logger,
	})
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
