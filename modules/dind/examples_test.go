package dind_test

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/testcontainers/testcontainers-go/modules/dind"
	"log"

	"github.com/testcontainers/testcontainers-go"
)

func ExampleRun() {
	// runDinDContainer {
	ctx := context.Background()

	dindContainer, err := dind.Run(ctx, "docker:28.0.1-dind")
	defer func() {
		if err := testcontainers.TerminateContainer(dindContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	// Retrieve the host where the DinD daemon is listening
	// didnHost {
	host, err := dindContainer.Host(ctx)
	if err != nil {
		log.Printf("failed to get docker host: %s", err)
		return
	}
	// }

	// getDockerClient {
	cli, err := client.NewClientWithOpts(client.WithHost(host), client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("failed to create docker client: %s", err)
		return
	}
	// }

	version, err := cli.ServerVersion(ctx)
	if err != nil {
		log.Printf("failed to get server version: %s", err)
		return
	}

	fmt.Println(version.APIVersion)

	// Output:
	// 1.48
}
