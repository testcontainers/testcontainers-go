package milvus_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/milvus"
)

func TestMilvus(t *testing.T) {
	ctr, err := milvus.Run(t.Context(), "milvusdb/milvus:v2.6.3")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("Connection endpoint", func(tt *testing.T) {
		// connectionString {
		ctx := tt.Context()
		connectionStr, err := ctr.ConnectionString(ctx)
		// }
		require.NoError(tt, err)

		require.NotEmpty(tt, connectionStr)
	})
}
