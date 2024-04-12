package openfga_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openfga"
)

func TestOpenFGA(t *testing.T) {
	ctx := context.Background()

	container, err := openfga.RunContainer(ctx, testcontainers.WithImage("openfga/openfga:v1.5.0"))
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
