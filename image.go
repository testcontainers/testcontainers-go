package testcontainers

import (
	"context"

	"github.com/testcontainers/testcontainers-go/image"
)

// Deprecated: use testcontainers-go [image.Info] instead
// ImageInfo represents summary information of an image
type ImageInfo = image.Info

// ImageProvider allows manipulating images
type ImageProvider interface {
	ListImages(context.Context) ([]ImageInfo, error) // Deprecated: use testcontainers-go [image.List] instead
	SaveImages(context.Context, string, ...string) error
	PullImage(context.Context, string) error
}
