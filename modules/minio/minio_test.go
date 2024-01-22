package minio

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

func TestMinio(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, testcontainers.WithImage("minio/minio:RELEASE.2024-01-16T16-07-38Z"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
}
