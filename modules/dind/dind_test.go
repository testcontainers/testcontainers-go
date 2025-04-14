package dind_test

import (
	"context"
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
	err = provider.PullImage(ctx, "nginx:1.27")
	require.NoError(t, err)

	t.Run("not-available", func(t *testing.T) {
		err := dindContainer.LoadImage(ctx, "fake.registry/fake:non-existing")
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		err := dindContainer.LoadImage(ctx, "nginx:1.27")
		require.NoError(t, err)

		images, err := cli.ImageList(ctx, image.ListOptions{})
		require.NoError(t, err)

		if len(images) == 0 || len(images) > 1 {
			t.Fatalf("got %d images, expected 1", len(images))
		}

		img, err := cli.ImageInspect(ctx, images[0].ID)
		require.NoError(t, err)

		require.Equal(t, "nginx:1.27", img.RepoTags[0])
		require.Equal(t, []string{"/docker-entrypoint.sh"}, []string(img.Config.Entrypoint))
		require.Equal(t, []string{"nginx", "-g", "daemon off;"}, []string(img.Config.Cmd))
	})
}
