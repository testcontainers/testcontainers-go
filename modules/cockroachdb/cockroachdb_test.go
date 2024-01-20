package cockroachdb

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

func TestCockroachDB(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, testcontainers.WithImage("cockroachdb/cockroach:latest-v23.1"))
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
