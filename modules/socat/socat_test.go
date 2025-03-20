package socat_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/socat"
)

func TestSocat(t *testing.T) {
	ctx := context.Background()

	ctr, err := socat.Run(ctx, "alpine/socat:1.8.0.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}
