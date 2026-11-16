package grafanalgtm_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	grafanalgtm "github.com/testcontainers/testcontainers-go/modules/grafana-lgtm"
)

func ExampleRun() {
	// runGrafanaLGTMContainer {
	ctx := context.Background()

	grafanaLgtmContainer, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0")
	defer func() {
		if err := testcontainers.TerminateContainer(grafanaLgtmContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := grafanaLgtmContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
