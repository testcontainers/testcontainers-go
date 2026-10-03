package sftp_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/sftp"
)

func TestSFTP(t *testing.T) {
	ctx := context.Background()

	ctr, err := sftp.Run(ctx, "atmoz/sftp:latest",
		sftp.WithUser("testuser", "testpassword"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("address", func(t *testing.T) {
		addr, err := ctr.Address(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, addr)
	})
}

func TestSFTP_NoUsers(t *testing.T) {
	ctx := context.Background()

	ctr, err := sftp.Run(ctx, "atmoz/sftp:latest")
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
	require.Nil(t, ctr)
	require.Contains(t, err.Error(), "at least one user is required")
}
