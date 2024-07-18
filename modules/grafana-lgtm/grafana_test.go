package grafanalgtm_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/grafanalgtm"
)

func TestGrafana(t *testing.T) {
	ctx := context.Background()

	container, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0")
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
