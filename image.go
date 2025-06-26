package testcontainers

import (
	"context"

	"github.com/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// ImageInfo represents summary information of an image
type ImageInfo struct {
	ID   string
	Name string
}

type saveOptions struct {
	Platforms []specs.Platform
}

type SaveImageOption func(*saveOptions) error

// SaveImageWithPlatforms allows specifying which platform(s) to save
func SaveImageWithPlatforms(plaforms ...string) SaveImageOption {
	return func(opts *saveOptions) error {
		p, err := platforms.ParseAll(plaforms)
		if err != nil {
			return err
		}
		opts.Platforms = p
		return nil
	}
}

// ImageProvider allows manipulating images
type ImageProvider interface {
	ListImages(context.Context) ([]ImageInfo, error)
	SaveImages(context.Context, string, ...string) error
	SaveImagesWithOps(context.Context, string, []string, ...SaveImageOption) error
	PullImage(context.Context, string) error
}
