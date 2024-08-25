package vearch_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/vearch"
)

func TestVearch(t *testing.T) {
	ctx := context.Background()

	container, err := vearch.Run(ctx, "vearch/vearch:3.5.1")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	t.Run("REST Endpoint", func(tt *testing.T) {
		// restEndpoint {
		restEndpoint, err := container.RESTEndpoint(ctx)
		// }
		if err != nil {
			tt.Fatalf("failed to get REST endpoint: %s", err)
		}

		cli := &http.Client{}
		resp, err := cli.Get(restEndpoint)
		if err != nil {
			tt.Fatalf("failed to perform GET request: %s", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			tt.Fatalf("unexpected status code: %d", resp.StatusCode)
		}
	})
}
