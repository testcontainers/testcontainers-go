package ravendb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/ravendb"
)

func ExampleRun() {
	// runRavenDBContainer {
	ctx := context.Background()

	ravendbContainer, err := ravendb.Run(ctx, "ravendb/ravendb:6.0-ubuntu-latest")
	defer func() {
		if err := testcontainers.TerminateContainer(ravendbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := ravendbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_managementURL() {
	ctx := context.Background()

	ravendbContainer, err := ravendb.Run(ctx, "ravendb/ravendb:6.0-ubuntu-latest")
	defer func() {
		if err := testcontainers.TerminateContainer(ravendbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	managementURL, err := ravendbContainer.ManagementURL(ctx)
	if err != nil {
		log.Printf("failed to get management URL: %s", err)
		return
	}

	isHTTP := len(managementURL) > 7 && managementURL[:7] == "http://"
	fmt.Println(isHTTP)

	// Output:
	// true
}
