package milvus_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/milvus"
)

func TestMilvus(t *testing.T) {
	ctx := context.Background()

	container, err := milvus.RunContainer(ctx, testcontainers.WithImage("milvusdb/milvus:v2.3.9"))
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
