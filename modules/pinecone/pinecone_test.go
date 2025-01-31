package pinecone_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/pinecone"
)

func TestPinecone(t *testing.T) {
	ctx := context.Background()

	ctr, err := pinecone.Run(ctx, "ghcr.io/pinecone-io/pinecone-local:latest")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}
