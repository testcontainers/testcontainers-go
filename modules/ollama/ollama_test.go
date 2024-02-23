package ollama_test

import (
	"context"
	"net/http"
	"strings"
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

	t.Run("ConnectionString", func(t *testing.T) {
		// connectionString {
		connectionStr, err := container.ConnectionString(ctx)
		// }
		if err != nil {
			t.Fatal(err)
		}

		httpClient := &http.Client{}
		resp, err := httpClient.Get(connectionStr)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code 200, got %d", resp.StatusCode)
		}
	})
}

func TestRunContainer_withModel_error(t *testing.T) {
	ctx := context.Background()

	ollamaContainer, err := ollama.RunContainer(
		ctx,
		testcontainers.WithImage("ollama/ollama:0.1.25"),
		ollama.WithModel("non-existent"),
	)
	if ollamaContainer != nil {
		t.Fatal("container should not start")
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErrorMessages := []string{
		"failed to run non-existent model",
		"Error: pull model manifest: file does not exist",
	}

	for _, expected := range expectedErrorMessages {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("expected error to contain %q, got %s", expected, err)
		}
	}
}
