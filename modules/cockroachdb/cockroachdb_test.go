package cockroachdb_test

import (
	"context"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

const testImage = "cockroachdb/cockroach:latest-v23.1"

func TestRun(t *testing.T) {
	testContainer(t, false)
}

func TestRun_WithAllOptions(t *testing.T) {
	testContainer(t, false,
		cockroachdb.WithDatabase("testDatabase"),
		cockroachdb.WithStoreSize("50%"),
		cockroachdb.WithUser("testUser"),
		cockroachdb.WithPassword("testPassword"),
		cockroachdb.WithNoClusterDefaults(),
		cockroachdb.WithInitScripts("testdata/__init.sql"),
		// WithInsecure is not present as it is incompatible with WithPassword.
	)
}

func TestRun_WithInsecure(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		testContainer(t, true, cockroachdb.WithInsecure())
	})

	t.Run("invalid-password-insecure", func(t *testing.T) {
		_, err := cockroachdb.Run(context.Background(), testImage,
			cockroachdb.WithPassword("testPassword"),
			cockroachdb.WithInsecure(),
		)
		require.Error(t, err)
	})

	t.Run("invalid-insecure-password", func(t *testing.T) {
		_, err := cockroachdb.Run(context.Background(), testImage,
			cockroachdb.WithInsecure(),
			cockroachdb.WithPassword("testPassword"),
		)
		require.Error(t, err)
	})
}

// testContainer runs a CockroachDB container and validates its functionality.
func testContainer(t *testing.T, insecure bool, opts ...testcontainers.ContainerCustomizer) {
	t.Helper()

	ctx := context.Background()
	ctr, err := cockroachdb.Run(ctx, testImage, opts...)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.NotNil(t, ctr)

	info, err := ctr.ConnectionInfo(ctx)
	require.NoError(t, err)
	u, err := url.Parse(info.URL)
	require.NoError(t, err)
	require.Equal(t, "postgres", u.Scheme)
	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	port, err := ctr.MappedPort(ctx, "26257/tcp")
	require.NoError(t, err)
	require.Equal(t, net.JoinHostPort(host, port.Port()), u.Host)
	if insecure {
		require.Equal(t, "disable", u.Query().Get("sslmode"))
		require.Nil(t, info.TLSConfig)
		_, err := ctr.TLSConfig()
		require.ErrorIs(t, err, cockroachdb.ErrTLSNotEnabled)
	} else {
		require.Equal(t, "verify-full", u.Query().Get("sslmode"))
		require.NotNil(t, info.TLSConfig)
		require.Equal(t, "127.0.0.1", info.TLSConfig.ServerName)
		require.NotNil(t, info.TLSConfig.RootCAs)
		require.Len(t, info.TLSConfig.Certificates, 1)
	}
	addr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.Equal(t, info.URL, addr)
	require.Equal(t, addr, ctr.MustConnectionString(ctx))
}
