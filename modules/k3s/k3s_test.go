package k3s_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/containerd/platforms"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_LoadImages(t *testing.T) {
	// Give up to three minutes to run this test
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	// runK3sContainer {
	k3sContainer, err := k3s.Run(ctx, "rancher/k3s:v1.27.1-k3s1")
	testcontainers.CleanupContainer(t, k3sContainer)
	require.NoError(t, err)
	// }

	// GetKubeConfig {
	kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, kubeConfigYaml)
	// }

	provider, err := testcontainers.ProviderDocker.GetProvider()
	require.NoError(t, err)

	// This function only works for single architecture images
	// Forces the test to use a single-arch version of the image
	arch := platforms.DefaultSpec().Architecture
	if platforms.DefaultSpec().Variant != "" {
		arch += platforms.DefaultSpec().Variant
	}
	nginxImg := arch + "/nginx"

	// ensure nginx image is available locally
	err = provider.PullImage(ctx, nginxImg)
	require.NoError(t, err)

	t.Run("Test load image not available", func(t *testing.T) {
		err := k3sContainer.LoadImages(ctx, "fake.registry/fake:non-existing")
		require.Error(t, err)
	})

	t.Run("Test load image in cluster", func(t *testing.T) {
		err := k3sContainer.LoadImages(ctx, nginxImg)
		require.NoError(t, err)

		// Pod creation and Running assertions are exercised in the child module.
	})
}

func Test_LoadImagesWithPlatform(t *testing.T) {
	// Give up to three minutes to run this test
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	k3sContainer, err := k3s.Run(ctx, "rancher/k3s:v1.27.1-k3s1")
	testcontainers.CleanupContainer(t, k3sContainer)
	require.NoError(t, err)

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, kubeConfigYaml)

	provider, err := testcontainers.ProviderDocker.GetProvider()
	require.NoError(t, err)

	// This function only works for single architecture images.
	// Forces the test to use a single-arch version of the image.
	hostPlatform := platforms.DefaultSpec()
	hostPlatform.OS = "linux"
	arch := hostPlatform.Architecture
	if hostPlatform.Variant != "" {
		arch += hostPlatform.Variant
	}
	nginxImg := arch + "/nginx"

	// ensure nginx image is available locally for the host platform
	err = provider.PullImageWithOpts(ctx, nginxImg, testcontainers.PullDockerImageWithPlatform(hostPlatform))
	require.NoError(t, err)

	t.Run("Test load image not available", func(t *testing.T) {
		p, _ := platforms.Parse("linux/amd64")
		err := k3sContainer.LoadImagesWithPlatform(ctx, []string{"fake.registry/fake:non-existing"}, &p)
		require.Error(t, err)
	})

	t.Run("Test load image with wrong architecture", func(t *testing.T) {
		pullPlatform, _ := platforms.Parse("linux/s390x")
		// A multi-platform index can reuse amd64 content already in containerd's
		// image store (for example from amd64/nginx), even after an s390x pull.
		// Use a single-architecture reference so amd64 is never an export target.
		img := "s390x/nginx:mainline"
		err = provider.PullImageWithOpts(ctx, img, testcontainers.PullDockerImageWithPlatform(pullPlatform))
		require.NoError(t, err)

		loadPlatform, _ := platforms.Parse("linux/amd64")
		err := k3sContainer.LoadImagesWithPlatform(ctx, []string{img}, &loadPlatform)
		require.Error(t, err)
		requirePlatformMismatchError(t, err, img, platforms.Format(loadPlatform))
	})

	t.Run("Test load image in cluster", func(t *testing.T) {
		err := k3sContainer.LoadImagesWithPlatform(ctx, []string{nginxImg}, &hostPlatform)
		require.NoError(t, err)

		// Pod creation and Running assertions are exercised in the child module.
	})
}

func requirePlatformMismatchError(t *testing.T, err error, img, platform string) {
	t.Helper()

	errMsg := err.Error()
	legacy := fmt.Sprintf(
		"image with reference %s was found but does not provide the specified platform (%s)",
		img,
		platform,
	)
	modern := "no suitable export target found for platform " + platform

	require.True(t,
		strings.Contains(errMsg, legacy) || strings.Contains(errMsg, modern),
		"unexpected platform mismatch error: %q", errMsg,
	)
}

func Test_APIServerReady(t *testing.T) {
	ctx := context.Background()

	k3sContainer, err := k3s.Run(ctx, "rancher/k3s:v1.27.1-k3s1")
	testcontainers.CleanupContainer(t, k3sContainer)
	require.NoError(t, err)

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, kubeConfigYaml)

	endpoint, err := k3sContainer.PortEndpoint(ctx, "6443/tcp", "https")
	require.NoError(t, err)
	require.Contains(t, string(kubeConfigYaml), endpoint)
}

func Test_WithManifestOption(t *testing.T) {
	ctx := context.Background()

	k3sContainer, err := k3s.Run(ctx,
		"rancher/k3s:v1.27.1-k3s1",
		k3s.WithManifest("nginx-manifest.yaml"),
		testcontainers.WithAdditionalWaitStrategy(wait.ForExec([]string{"kubectl", "wait", "pod", "nginx", "--for=condition=Ready"})),
	)
	testcontainers.CleanupContainer(t, k3sContainer)
	require.NoError(t, err)
}
