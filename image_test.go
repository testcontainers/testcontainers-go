package testcontainers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/containerd/platforms"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

func TestImageList(t *testing.T) {
	t.Setenv("DOCKER_HOST", core.MustExtractDockerHost(context.Background()))

	provider, err := ProviderDocker.GetProvider()
	require.NoErrorf(t, err, "failed to get provider")

	defer func() {
		err = provider.Close()
		require.NoError(t, err)
	}()

	req := ContainerRequest{
		Image: "redis:latest",
	}

	ctr, err := provider.CreateContainer(context.Background(), req)
	CleanupContainer(t, ctr)
	require.NoErrorf(t, err, "creating test container")

	images, err := provider.ListImages(context.Background())
	require.NoErrorf(t, err, "listing images")

	require.NotEmptyf(t, images, "no images retrieved")

	// look if the list contains the container image
	for _, img := range images {
		if img.Name == req.Image {
			return
		}
	}

	t.Fatalf("expected image not found: %s", req.Image)
}

func TestSaveImages(t *testing.T) {
	t.Setenv("DOCKER_HOST", core.MustExtractDockerHost(context.Background()))

	provider, err := ProviderDocker.GetProvider()
	require.NoErrorf(t, err, "failed to get provider")

	defer func() {
		err = provider.Close()
		require.NoError(t, err)
	}()

	req := ContainerRequest{
		Image: "redis:latest",
	}

	ctr, err := provider.CreateContainer(context.Background(), req)
	CleanupContainer(t, ctr)
	require.NoErrorf(t, err, "creating test container")

	output := filepath.Join(t.TempDir(), "images.tar")
	err = provider.SaveImages(context.Background(), output, req.Image)
	require.NoErrorf(t, err, "saving image %q", req.Image)

	info, err := os.Stat(output)
	require.NoError(t, err)

	require.NotZerof(t, info.Size(), "output file is empty")
}

func TestSaveImagesWithOpts(t *testing.T) {
	t.Setenv("DOCKER_HOST", core.MustExtractDockerHost(context.Background()))

	provider, err := ProviderDocker.GetProvider()
	require.NoErrorf(t, err, "failed to get provider")

	defer func() {
		err = provider.Close()
		require.NoError(t, err)
	}()

	req := ContainerRequest{
		Image:         "redis:latest",
		ImagePlatform: "linux/amd64",
	}

	p, err := platforms.ParseAll([]string{"linux/amd64"})
	require.NoError(t, err)

	ctr, err := provider.CreateContainer(context.Background(), req)
	CleanupContainer(t, ctr)
	require.NoErrorf(t, err, "creating test container")

	output := filepath.Join(t.TempDir(), "images.tar")
	err = provider.SaveImagesWithOpts(
		context.Background(), output, []string{req.Image}, SaveDockerImageWithPlatforms(p...),
	)
	require.NoErrorf(t, err, "saving image %q", req.Image)

	info, err := os.Stat(output)
	require.NoError(t, err)

	require.NotZerof(t, info.Size(), "output file is empty")
}
