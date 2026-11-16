package openfga_test

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openfga"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ExampleRun() {
	// runOpenFGAContainer {
	ctx := context.Background()

	openfgaContainer, err := openfga.Run(ctx, "openfga/openfga:v1.18.0")
	defer func() {
		if err := testcontainers.TerminateContainer(openfgaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := openfgaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connectToPlayground() {
	// enablePlayground {
	openfgaContainer, err := openfga.Run(
		context.Background(),
		"openfga/openfga:v1.18.0",
		testcontainers.WithEnv(map[string]string{
			"OPENFGA_PLAYGROUND_ENABLED": "true",
			"OPENFGA_PLAYGROUND_ADDR":    "0.0.0.0:3000",
		}),
		testcontainers.WithAdditionalWaitStrategy(
			wait.ForHTTP("/playground").WithPort("3000/tcp").WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			})),
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(openfgaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// playgroundEndpoint {
	playgroundEndpoint, err := openfgaContainer.PlaygroundEndpoint(context.Background())
	if err != nil {
		log.Printf("failed to get playground endpoint: %s", err)
		return
	}
	// }

	httpClient := http.Client{}

	resp, err := httpClient.Get(playgroundEndpoint)
	if err != nil {
		log.Printf("failed to get playground endpoint: %s", err)
		return
	}

	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}
