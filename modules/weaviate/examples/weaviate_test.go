package weaviate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/weaviate"
)

func TestWeaviate(t *testing.T) {
	ctx := context.Background()

	ctr, err := weaviate.Run(ctx, "semitechnologies/weaviate:1.29.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("GrpcHostAddress", func(t *testing.T) {
		grpcHost, err := ctr.GrpcHostAddress(ctx)
		require.NoError(t, err)

		// grpcHealthCheck {
		opts := make([]grpc.DialOption, 0, 1)
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		conn, err := grpc.NewClient(grpcHost, opts...)
		require.NoErrorf(t, err, "failed to dial connection")
		defer conn.Close()
		client := grpc_health_v1.NewHealthClient(conn)
		check, err := client.Check(context.TODO(), &grpc_health_v1.HealthCheckRequest{})
		require.NoErrorf(t, err, "failed to get a health check")
		require.Equalf(t, grpc_health_v1.HealthCheckResponse_SERVING.Enum().Number(), check.Status.Number(), "unexpected status code: %d", check.Status.Number())
		// }
	})
}
