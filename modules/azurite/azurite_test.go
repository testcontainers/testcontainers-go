package azurite_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/azurite"
)

func TestAzurite(t *testing.T) {
	ctx := context.Background()

	container, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.23.0")
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
