package testcontainers

import (
	"context"

	"github.com/moby/moby/client"
)

// ImageInfo represents summary information of an image
type ImageInfo struct {
	ID   string
	Name string
}

type saveImageOptions struct {
	dockerSaveOpts []client.ImageSaveOption
}

type SaveImageOption func(*saveImageOptions) error

type pullImageOptions struct {
	dockerPullOpts client.ImagePullOptions
}

type PullImageOption func(*pullImageOptions) error

// ImageProvider allows manipulating images
type ImageProvider interface {
	ListImages(context.Context) ([]ImageInfo, error)
	SaveImages(context.Context, string, ...string) error
	SaveImagesWithOpts(context.Context, string, []string, ...SaveImageOption) error
	PullImage(context.Context, string) error
	PullImageWithPlatform(context.Context, string, string) error
}
