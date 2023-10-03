package k6

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

func Testk6(t *testing.T) {
	ctx := context.Background()

	container, err := runContainer(ctx, testcontainers.WithImage("grafana/k6"))
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
