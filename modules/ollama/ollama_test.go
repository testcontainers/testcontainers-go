package ollama_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/ollama"
)

func TestOllama(t *testing.T) {
	ctx := context.Background()

	container, err := ollama.RunContainer(ctx, testcontainers.WithImage("ollama/ollama:0.1.25"))
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
