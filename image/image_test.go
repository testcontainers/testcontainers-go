package image

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types/image"

	"github.com/testcontainers/testcontainers-go/internal/core"
	tclog "github.com/testcontainers/testcontainers-go/log"
)

func TestList(t *testing.T) {
	t.Setenv("DOCKER_HOST", core.MustExtractDockerHost(context.Background()))

	imageName := "redis:latest"

	err := Pull(context.Background(), imageName, tclog.StandardLogger(), image.PullOptions{})
	if err != nil {
		t.Fatalf("pulling image %v", err)
	}

	images, err := List(context.Background())
	if err != nil {
		t.Fatalf("listing images %v", err)
	}

	if len(images) == 0 {
		t.Fatal("no images retrieved")
	}

	// look if the list contains the container image
	for _, img := range images {
		if img.Name == imageName {
			return
		}
	}

	t.Fatalf("expected image not found: %s", imageName)
}

func TestSave(t *testing.T) {
	t.Setenv("DOCKER_HOST", core.MustExtractDockerHost(context.Background()))

	imageName := "redis:latest"

	err := Pull(context.Background(), imageName, tclog.StandardLogger(), image.PullOptions{})
	if err != nil {
		t.Fatalf("pulling image %v", err)
	}

	output := filepath.Join(t.TempDir(), "images.tar")
	err = SaveImages(context.Background(), output, imageName)
	if err != nil {
		t.Fatalf("saving image %q: %v", imageName, err)
	}

	info, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}

	if info.Size() == 0 {
		t.Fatalf("output file is empty")
	}
}
