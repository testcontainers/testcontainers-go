package openfga_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/openfga"
)

func TestOpenFGA(t *testing.T) {
	ctx := context.Background()

	container, err := openfga.Run(ctx, "openfga/openfga:v1.5.0")
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
