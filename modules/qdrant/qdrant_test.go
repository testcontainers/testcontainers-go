package qdrant_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/qdrant"
)

func TestQdrant(t *testing.T) {
	ctx := context.Background()

	container, err := qdrant.RunContainer(ctx, testcontainers.WithImage("qdrant/qdrant:v1.7.4"))
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
