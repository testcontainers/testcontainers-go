package milvus_test

import (
	"context"
	"testing"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/milvus"
)

func TestMilvus(t *testing.T) {
	ctx := context.Background()

	ctr, err := milvus.Run(ctx, "milvusdb/milvus:v2.3.9")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("Connect to Milvus with gRPC", func(tt *testing.T) {
		// connectionString {
		connectionStr, err := ctr.ConnectionString(ctx)
		// }
		require.NoError(t, err)

		milvusClient, err := client.NewGrpcClient(context.Background(), connectionStr)
		require.NoError(t, err)

		defer milvusClient.Close()

		v, err := milvusClient.GetVersion(ctx)
		require.NoError(t, err)

		tt.Logf("Milvus version: %s", v)
	})
}
