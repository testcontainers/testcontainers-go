package pinecone_test

import (
	"context"
	"testing"

	"github.com/pinecone-io/go-pinecone/v2/pinecone"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcpinecone "github.com/testcontainers/testcontainers-go/modules/pinecone"
)

func TestPinecone(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcpinecone.Run(ctx, "ghcr.io/pinecone-io/pinecone-local:v0.7.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	host, err := ctr.HttpEndpoint()
	require.NoError(t, err)

	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: "testcontainers-go", // API key is required, else use headers
		Host:   host,
	})
	require.NoError(t, err)

	indexes, err := pc.ListIndexes(ctx)
	require.NoError(t, err)
	require.Empty(t, indexes)
}

func TestPinecone_index(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcpinecone.Run(ctx, "ghcr.io/pinecone-io/pinecone-local:v0.7.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	host, err := ctr.HttpEndpoint()
	require.NoError(t, err)

	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: "testcontainers-go", // API key is required, else use headers
		Host:   host,
	})
	require.NoError(t, err)

	indexName := "my-serverless-index"

	_, err = pc.CreateServerlessIndex(ctx, &pinecone.CreateServerlessIndexRequest{
		Name:      indexName,
		Dimension: 3,
		Metric:    pinecone.Cosine,
		Cloud:     pinecone.Aws,
		Region:    "us-east-1",
		Tags:      &pinecone.IndexTags{"environment": "development"},
	})
	require.NoError(t, err)

	indexes, err := pc.ListIndexes(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, indexes)
	require.Equal(t, indexes[0].Name, indexName)
}
