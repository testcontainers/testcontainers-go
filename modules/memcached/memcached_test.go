package memcached_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/memcached"
)

func TestRun(t *testing.T) {
	ctx := t.Context()

	ctr, err := memcached.Run(ctx, "memcached:1.6-alpine")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}
