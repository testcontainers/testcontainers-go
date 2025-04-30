package valkey_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/valkey-io/valkey-go"

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

func ExampleRun_withTLS() {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx,
		"valkey/valkey:7.2.5",
		tcvalkey.WithSnapshotting(10, 1),
		tcvalkey.WithLogLevel(tcvalkey.LogLevelVerbose),
		tcvalkey.WithTLS(),
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

	if valkeyContainer.TLSConfig() == nil {
		log.Println("TLS is not enabled")
		return
	}

	uri, err := valkeyContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	// You will likely want to wrap your Valkey package of choice in an
	// interface to aid in unit testing and limit lock-in throughout your
	// codebase but that's out of scope for this example
	options, err := valkey.ParseURL(uri)
	if err != nil {
		log.Printf("failed to parse connection string: %s", err)
		return
	}

	options.TLSConfig = valkeyContainer.TLSConfig()

	client, err := valkey.NewClient(options)
	if err != nil {
		log.Printf("failed to create valkey client: %s", err)
		return
	}
	defer func() {
		err := flushValkey(ctx, client)
		if err != nil {
			log.Printf("failed to flush valkey: %s", err)
		}
	}()

	resp := client.Do(ctx, client.B().Ping().Build().Pin())
	fmt.Println(resp.String())

	// Output:
	// {"Message":{"Value":"PONG","Type":"simple string"}}
}
