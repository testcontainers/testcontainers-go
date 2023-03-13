package vault

import (
	"context"
	"testing"
)

func TestVault(t *testing.T) {
	ctx := context.Background()

	container, err := StartContainer(ctx)
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
