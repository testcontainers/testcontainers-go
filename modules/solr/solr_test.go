package solr_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/solr"
)

func TestSolr(t *testing.T) {
	ctx := context.Background()

	ctr, err := solr.Run(ctx, "solr:9")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("Address", func(t *testing.T) {
		addr, err := ctr.Address(ctx)
		require.NoError(t, err)
		require.Contains(t, addr, "http://")
		require.Contains(t, addr, "/solr")
	})

	t.Run("CollectionURL", func(t *testing.T) {
		url, err := ctr.CollectionURL(ctx, "myCollection")
		require.NoError(t, err)
		require.Contains(t, url, "/solr/myCollection")
	})
}

func TestSolrWithCollection(t *testing.T) {
	ctx := context.Background()

	const collectionName = "testcollection"

	ctr, err := solr.Run(ctx, "solr:9",
		solr.WithCollection(collectionName),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	url, err := ctr.CollectionURL(ctx, collectionName)
	require.NoError(t, err)
	require.Contains(t, url, "/solr/"+collectionName)
}
