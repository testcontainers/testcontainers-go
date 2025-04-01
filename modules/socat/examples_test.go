package socat_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/socat"
	"github.com/testcontainers/testcontainers-go/network"
)

func ExampleRun() {
	ctx := context.Background()

	// createNetwork {
	nw, err := network.New(ctx)
	if err != nil {
		log.Printf("failed to create network: %v", err)
		return
	}
	defer func() {
		if err := nw.Remove(ctx); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()
	// }

	// createHelloWorldContainer {
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "testcontainers/helloworld:1.2.0",
			ExposedPorts: []string{"8080/tcp"},
			Networks:     []string{nw.Name},
			NetworkAliases: map[string][]string{
				nw.Name: {"helloworld"},
			},
		},
		Started: true,
	})
	if err != nil {
		log.Printf("failed to create container: %v", err)
		return
	}
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	// }

	// createSocatContainer {
	const exposedPort = 8080

	target := socat.NewTarget(exposedPort, "helloworld")

	socatContainer, err := socat.Run(
		ctx, "alpine/socat:1.8.0.1",
		socat.WithTarget(target),
		network.WithNetwork([]string{"socat"}, nw),
	)
	if err != nil {
		log.Printf("failed to create container: %v", err)
		return
	}
	defer func() {
		if err := testcontainers.TerminateContainer(socatContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	// }

	// readFromSocat {
	httpClient := http.DefaultClient

	baseURI := socatContainer.TargetURL(exposedPort)

	resp, err := httpClient.Get(baseURI.String() + "/ping")
	if err != nil {
		log.Printf("failed to get response: %v", err)
		return
	}
	defer resp.Body.Close()
	// }

	fmt.Printf("%d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read body: %v", err)
		return
	}

	fmt.Printf("%s", string(body))

	// Output:
	// 200
	// PONG
}
