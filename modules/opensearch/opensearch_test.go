package opensearch_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/opensearch"
)

func TestOpenSearch(t *testing.T) {
	ctx := context.Background()

	container, err := opensearch.RunContainer(ctx, testcontainers.WithImage("opensearchproject/opensearch:2.11.1"))
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
