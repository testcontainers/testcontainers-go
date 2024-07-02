package qdrant_test

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go/modules/qdrant"
)

func TestQdrant(t *testing.T) {
	ctx := context.Background()

	container, err := qdrant.Run(ctx, "qdrant/qdrant:v1.7.4")
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

	t.Run("gRPC Endpoint", func(tt *testing.T) {
		// gRPCEndpoint {
		grpcEndpoint, err := container.GRPCEndpoint(ctx)
		// }
		if err != nil {
			tt.Fatalf("failed to get REST endpoint: %s", err)
		}

		conn, err := grpc.Dial(grpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
	})

	t.Run("Web UI", func(tt *testing.T) {
		// webUIEndpoint {
		webUI, err := container.WebUI(ctx)
		// }
		if err != nil {
			tt.Fatalf("failed to get REST endpoint: %s", err)
		}

		cli := &http.Client{}
		resp, err := cli.Get(webUI)
		if err != nil {
			tt.Fatalf("failed to perform GET request: %s", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			tt.Fatalf("unexpected status code: %d", resp.StatusCode)
		}
	})
}
