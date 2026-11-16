package minio_test

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

func TestMinio(t *testing.T) {
	ctx := context.Background()

	ctr, err := tcminio.Run(ctx,
		"quay.io/minio/minio:RELEASE.2024-01-16T16-07-38Z",
		tcminio.WithUsername("thisismyuser"), tcminio.WithPassword("thisismypassword"))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	// connectionString {
	url, err := ctr.ConnectionString(ctx)
	// }
	require.NoError(t, err)

	require.Equal(t, "thisismyuser", ctr.Username)
	require.Equal(t, "thisismypassword", ctr.Password)
	host, port, err := net.SplitHostPort(url)
	require.NoError(t, err)
	require.NotEmpty(t, host)
	require.NotEmpty(t, port)
	// SDK bucket creation and full object upload/read coverage live in examples/minio_test.go.
}
