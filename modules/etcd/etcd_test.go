package etcd_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/etcd"
)

func Testetcd(t *testing.T) {
	ctx := context.Background()

	ctr, err := etcd.Run(ctx, "bitnami/etcd:latest")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}
