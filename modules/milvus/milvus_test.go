package milvus_test

import (
	"testing"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/milvus"
)

func TestMilvus(t *testing.T) {
	ctr, err := milvus.Run(t.Context(), "milvusdb/milvus:v2.3.9")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("Connect to Milvus with gRPC", func(tt *testing.T) {
		// connectionString {
		ctx := tt.Context()
		connectionStr, err := ctr.ConnectionString(ctx)
		// }
		require.NoError(t, err)

		milvusClient, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
			Address: connectionStr,
		})
		require.NoError(t, err)

		defer milvusClient.Close(t.Context())

		v, err := milvusClient.GetServerVersion(ctx, milvusclient.NewGetServerVersionOption())
		require.NoError(t, err)

		tt.Logf("Milvus version: %s", v)
	})
}
