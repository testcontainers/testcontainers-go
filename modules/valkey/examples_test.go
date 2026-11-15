package valkey_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	tcvalkey "github.com/testcontainers/testcontainers-go/modules/valkey"
)

func ExampleRun() {
	// runValkeyContainer {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx,
		"valkey/valkey:7.2.5",
		tcvalkey.WithSnapshotting(10, 1),
		tcvalkey.WithLogLevel(tcvalkey.LogLevelVerbose),
		tcvalkey.WithConfigFile(filepath.Join("testdata", "valkey7.conf")),
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
