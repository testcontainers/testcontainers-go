package dind_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dind"
)

func Test_LoadImages(t *testing.T) {
	// Give up to three minutes to run this test
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	dindContainer, err := dind.Run(ctx, "docker:28.0.1-dind")
	testcontainers.CleanupContainer(t, dindContainer)
	require.NoError(t, err)

	host, err := dindContainer.Host(ctx)
	require.NoError(t, err)

	cli, err := client.NewClientWithOpts(client.WithHost(host), client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	provider, err := testcontainers.ProviderDocker.GetProvider()
	require.NoError(t, err)

	// ensure nginx image is available locally
	err = provider.PullImage(ctx, "nginx")
	require.NoError(t, err)

	t.Run("not-available", func(t *testing.T) {
		err := dindContainer.LoadImage(ctx, "fake.registry/fake:non-existing")
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		err := dindContainer.LoadImage(ctx, "nginx")
		require.NoError(t, err)

		images, err := cli.ImageList(ctx, image.ListOptions{})
		require.NoError(t, err)

		found := slices.ContainsFunc(images, func(img image.Summary) bool {
			return len(img.RepoTags) > 0 && img.RepoTags[0] == "nginx:latest"
		})

		require.True(t, found)
	})
}
