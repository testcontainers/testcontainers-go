// Package fakegcsserver_test provides integration tests and usage examples
// for the fakegcsserver Testcontainers module.
package fakegcsserver_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/fakegcsserver"
)

// ExampleRun demonstrates how to start a FakeGCSServer container and
// retrieve its GCS-compatible storage URL.
func ExampleRun() {
	// runFakeGCSServerContainer {
	ctx := context.Background()

	fakegcsserverContainer, err := fakegcsserver.Run(ctx, "fsouza/fake-gcs-server:1.47")
	defer func() {
		if err := testcontainers.TerminateContainer(fakegcsserverContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	storageURL, err := fakegcsserverContainer.StorageURL(ctx)
	if err != nil {
		log.Printf("failed to get storage URL: %s", err)
		return
	}

	fmt.Println(storageURL != "")

	// Output:
	// true
}
