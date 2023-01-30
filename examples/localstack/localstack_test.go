package localstack

import (
	"context"
	"testing"
)

func TestLocalStack(t *testing.T) {
	ctx := context.Background()

	container, err := setupLocalStack(ctx)
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
