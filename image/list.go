package image

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/image"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// List list images from the provider. If an image has multiple Tags, each tag is reported
// individually with the same ID and same labels
func List(ctx context.Context) ([]Info, error) {
	images := []Info{}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return images, err
	}
	defer cli.Close()

	imageList, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return images, fmt.Errorf("listing images %w", err)
	}

	for _, img := range imageList {
		for _, tag := range img.RepoTags {
			images = append(images, Info{ID: img.ID, Name: tag})
		}
	}

	return images, nil
}
