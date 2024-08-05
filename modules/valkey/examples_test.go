package valkey_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
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
	defer func() {
		if err := testcontainers.TerminateContainer(valkeyContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := valkeyContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
