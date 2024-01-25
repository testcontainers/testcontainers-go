package minio_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
)

func ExampleRunContainer() {
	// runMinioContainer {
	ctx := context.Background()

	minioContainer, err := minio.RunContainer(ctx, testcontainers.WithImage("minio/minio:RELEASE.2024-01-16T16-07-38Z"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := minioContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := minioContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
