package testcontainers

import (
	"context"
)

// DockerImage represents a Docker image as a string.
// Valid examples:
// - "alpine:3.12"
// - "docker.io/nginx:latest"
// - "my-registry.local:5000/my-image:my-tag"
type DockerImage string

func NewDockerImage(image string) DockerImage {
	return DockerImage(image)
}

func (d DockerImage) String() string {
	return string(d)
}

// ImageInfo represents a summary information of an image
type ImageInfo struct {
	ID   string
	Name string
}

// ImageProvider allows manipulating images
type ImageProvider interface {
	ListImages(context.Context) ([]ImageInfo, error)
	SaveImages(context.Context, string, ...string) error
	PullImage(context.Context, string) error
}
