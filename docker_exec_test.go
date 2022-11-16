package testcontainers

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExec(t *testing.T) {
	ctx := context.Background()
	req := ContainerRequest{
		Image: nginxAlpineImage,
	}

	container, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, container)

	code, reader, err := container.Exec(ctx, []string{"ls", "/usr/share/nginx"})
	require.NoError(t, err)
	require.Zero(t, code)
	require.NotNil(t, reader)

	b, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NotNil(t, b)

	str := string(b)
	require.Equal(t, "html\n", str)
}
