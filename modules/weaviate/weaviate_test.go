package weaviate_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	wvt "github.com/weaviate/weaviate-go-client/v4/weaviate"
	wvtgrpc "github.com/weaviate/weaviate-go-client/v4/weaviate/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/weaviate"
)

func TestWeaviate(t *testing.T) {
	ctx := context.Background()

	container, err := weaviate.RunContainer(ctx, testcontainers.WithImage("semitechnologies/weaviate:1.25.5"))
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

	t.Run("GrpcHostAddress", func(tt *testing.T) {
		// gRPCHostAddress {
		host, err := container.GrpcHostAddress(ctx)
		// }
		if err != nil {
			t.Fatal(err)
		}

		var opts []grpc.DialOption
		opts = append(opts, grpc.WithBlock())
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		conn, err := grpc.Dial(host, opts...)
		if err != nil {
			tt.Fatalf("failed to dial connection: %v", err)
		}
		client := grpc_health_v1.NewHealthClient(conn)
		check, err := client.Check(context.TODO(), &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			tt.Fatalf("failed to get a health check: %v", err)
		}
		if grpc_health_v1.HealthCheckResponse_SERVING.Enum().Number() != check.Status.Number() {
			tt.Fatalf("unexpected status code: %d", check.Status.Number())
		}
	})

	t.Run("Weaviate client", func(tt *testing.T) {
		httpScheme, httpHost, err := container.HttpHostAddress(ctx)
		if err != nil {
			tt.Fatal(err)
		}
		grpcHost, err := container.GrpcHostAddress(ctx)
		if err != nil {
			tt.Fatal(err)
		}
		config := wvt.Config{Scheme: httpScheme, Host: httpHost, GrpcConfig: &wvtgrpc.Config{Host: grpcHost}}
		client, err := wvt.NewClient(config)
		if err != nil {
			tt.Fatal(err)
		}
		meta, err := client.Misc().MetaGetter().Do(ctx)
		if err != nil {
			tt.Fatal(err)
		}

		if meta == nil || meta.Version == "" {
			tt.Fatal("failed to get /v1/meta response")
		}
	})
}
