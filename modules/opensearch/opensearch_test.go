package opensearch_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/opensearch"
)

func TestOpenSearch(t *testing.T) {
	ctx := context.Background()

	container, err := opensearch.RunContainer(ctx, testcontainers.WithImage("opensearchproject/opensearch:2.11.1"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	t.Run("Connect to Address", func(t *testing.T) {
		address, err := container.Address(ctx)
		if err != nil {
			t.Fatal(err)
		}

		client := &http.Client{}

		req, err := http.NewRequest("GET", address, nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to perform GET request: %s", err)
		}
		defer resp.Body.Close()
	})
}
