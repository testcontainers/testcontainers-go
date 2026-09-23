package nginx_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/nginx"
)

func ExampleRun() {
	// runNginxContainer {
	ctx := context.Background()

	nginxContainer, err := nginx.Run(ctx, "nginx:1.25")
	defer func() {
		if err := testcontainers.TerminateContainer(nginxContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := nginxContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_httpEndpoint() {
	// httpEndpointExample {
	ctx := context.Background()

	nginxContainer, err := nginx.Run(ctx, "nginx:1.25")
	defer func() {
		if err := testcontainers.TerminateContainer(nginxContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	httpEndpoint, err := nginxContainer.HTTPEndpoint(ctx)
	if err != nil {
		log.Printf("failed to get HTTP endpoint: %s", err)
		return
	}
	// }

	_ = httpEndpoint
	fmt.Println("got http endpoint")

	// Output:
	// got http endpoint
}
