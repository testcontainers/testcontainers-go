package minio_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go/modules/minio"
)

func ExampleRun() {
	// runMinioContainer {
	ctx := context.Background()

	minioContainer, err := minio.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := minioContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := minioContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
