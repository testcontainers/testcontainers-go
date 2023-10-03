package k6

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

func TestK6(t *testing.T) {
	ctx := context.Background()

	container, err := RunContainer(ctx, testcontainers.WithImage("szkiba/k6x"))
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
