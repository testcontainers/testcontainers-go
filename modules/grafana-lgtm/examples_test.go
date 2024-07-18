package grafanalgtm_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go/modules/grafanalgtm"
)

func ExampleRun() {
	// runGrafanaContainer {
	ctx := context.Background()

	grafanaContainer, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := grafanaContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := grafanaContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
