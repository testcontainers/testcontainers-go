package testcontainers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

func TestImageList(t *testing.T) {
	t.Setenv("DOCKER_HOST", core.MustExtractDockerHost(context.Background()))

	provider, err := ProviderDocker.GetProvider()
	if err != nil {
		t.Fatalf("failed to get provider %v", err)
	}

	defer func() {
		_ = provider.Close()
	}()

	req := ContainerRequest{
		Image: "redis:latest",
	}

	ctr, err := provider.CreateContainer(context.Background(), req)
	CleanupContainer(t, ctr)
	if err != nil {
		t.Fatalf("creating test container %v", err)
	}

	images, err := provider.ListImages(context.Background())
	if err != nil {
		t.Fatalf("listing images %v", err)
	}

	if len(images) == 0 {
		t.Fatal("no images retrieved")
	}

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
	if err != nil {
		t.Fatalf("failed to get provider %v", err)
	}

	defer func() {
		_ = provider.Close()
	}()

	req := ContainerRequest{
		Image: "redis:latest",
	}

	ctr, err := provider.CreateContainer(context.Background(), req)
	CleanupContainer(t, ctr)
	if err != nil {
		t.Fatalf("creating test container %v", err)
	}

	output := filepath.Join(t.TempDir(), "images.tar")
	err = provider.SaveImages(context.Background(), output, req.Image)
	if err != nil {
		t.Fatalf("saving image %q: %v", req.Image, err)
	}

	info, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}

	if info.Size() == 0 {
		t.Fatalf("output file is empty")
	}
}
