package weaviate_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/weaviate"
)

func TestWeaviate(t *testing.T) {
	ctx := context.Background()

	container, err := weaviate.RunContainer(ctx, testcontainers.WithImage("semitechnologies/weaviate:1.23.9"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	t.Run("HttpHostAddress", func(tt *testing.T) {
		// httpHostAddress {
		schema, host, err := container.HttpHostAddress(ctx)
		// }
		if err != nil {
			t.Fatal(err)
		}

		cli := &http.Client{}
		resp, err := cli.Get(fmt.Sprintf("%s://%s", schema, host))
		if err != nil {
			tt.Fatalf("failed to perform GET request: %s", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			tt.Fatalf("unexpected status code: %d", resp.StatusCode)
		}
	})
}
