package neo4j

import (
	"context"
	"fmt"
	"net/http"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	tclog "github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// containerPorts {
	defaultBoltPort  = "7687"
	defaultHttpPort  = "7474"
	defaultHttpsPort = "7473"
	// }
)

// Container represents the Neo4j container type used in the module
type Container struct {
	*testcontainers.DockerContainer
}

// BoltUrl returns the bolt url for the Neo4j container, using the bolt port, in the format of neo4j://host:port
func (c Container) BoltUrl(ctx context.Context) (string, error) {
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

// Run creates an instance of the Neo4j container type
func Run(ctx context.Context, img string, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	httpPort, _ := nat.NewPort("tcp", defaultHttpPort)
	req := testcontainers.Request{
		Image: img,
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
		Started: true,
		Logger:  tclog.StandardLogger(),
	}

	if len(opts) == 0 {
		opts = append(opts, WithoutAuthentication())
	}

	for _, option := range opts {
		if err := option.Customize(&req); err != nil {
			return nil, err
		}
	}

	err := validate(&req)
	if err != nil {
		return nil, err
	}

	ctr, err := testcontainers.Run(ctx, req)
	if err != nil {
		return nil, err
	}

	return &Container{DockerContainer: ctr}, nil
}

func isHttpOk() func(status int) bool {
	return func(status int) bool {
		return status == http.StatusOK
	}
}
