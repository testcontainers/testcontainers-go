package valkey_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go/modules/valkey"
)

func ExampleRun() {
	// runValkeyContainer {
	ctx := context.Background()

	valkeyContainer, err := valkey.Run(ctx,
		"docker.io/valkey/valkey:7.2.5",
		valkey.WithSnapshotting(10, 1),
		valkey.WithLogLevel(valkey.LogLevelVerbose),
		valkey.WithConfigFile(filepath.Join("testdata", "valkey7.conf")),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := valkeyContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := valkeyContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
