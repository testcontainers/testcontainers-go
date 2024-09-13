package toxiproxy

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// toxiproxyContainer represents the toxiproxy container type used in the module
type toxiproxyContainer struct {
	*testcontainers.DockerContainer
	URI string
}

// startContainer creates an instance of the toxiproxy container type
func startContainer(ctx context.Context, network string, networkAlias []string) (*toxiproxyContainer, error) {
	req := testcontainers.Request{
		Image:        "ghcr.io/shopify/toxiproxy:2.5.0",
		ExposedPorts: []string{"8474/tcp", "8666/tcp"},
		WaitingFor:   wait.ForHTTP("/version").WithPort("8474/tcp"),
		Networks: []string{
			network,
		},
		NetworkAliases: map[string][]string{
			network: networkAlias,
		},
		Started: true,
	}
	ctr, err := testcontainers.Run(ctx, req)
	var toxiC *toxiproxyContainer
	if ctr != nil {
		toxiC = &toxiproxyContainer{DockerContainer: ctr}
	}
	if err != nil {
		return toxiC, err
	}

	mappedPort, err := ctr.MappedPort(ctx, "8474")
	if err != nil {
		return toxiC, err
	}

	hostIP, err := ctr.Host(ctx)
	if err != nil {
		return toxiC, err
	}

	toxiC.URI = fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())

	return toxiC, nil
}
