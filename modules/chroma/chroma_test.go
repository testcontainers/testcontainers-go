package chroma_test

import (
	"context"
	"net/http"
	"testing"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/chroma"
)

func TestChroma(t *testing.T) {
	ctx := context.Background()

	container, err := chroma.RunContainer(ctx, testcontainers.WithImage("chromadb/chroma:0.4.24"))
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	t.Run("REST Endpoint retrieve docs site", func(tt *testing.T) {
		// restEndpoint {
		restEndpoint, err := container.RESTEndpoint(ctx)
		// }
		if err != nil {
			tt.Fatalf("failed to get REST endpoint: %s", err)
		}

		cli := &http.Client{}
		resp, err := cli.Get(restEndpoint + "/docs")
		if err != nil {
			tt.Fatalf("failed to perform GET request: %s", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			tt.Fatalf("unexpected status code: %d", resp.StatusCode)
		}
	})

	t.Run("GetClient", func(tt *testing.T) {
		// restEndpoint {
		endpoint, err := container.RESTEndpoint(context.Background())
		if err != nil {
			tt.Fatalf("failed to get REST endpoint: %s", err) // nolint:gocritic
		}
		chromaClient, err := chromago.NewClient(endpoint)
		// }
		if err != nil {
			tt.Fatalf("failed to create client: %s", err)
		}

		hb, err := chromaClient.Heartbeat(context.TODO())
		require.NoError(tt, err)
		require.NotNil(tt, hb["nanosecond heartbeat"])
	})
}
