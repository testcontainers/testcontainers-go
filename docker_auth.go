package testcontainers

import (
	"context"

	"github.com/docker/docker/api/types/registry"

	"github.com/testcontainers/testcontainers-go/auth"
)

// Deprecated: use auth.ForDockerImage instead
// DockerImageAuth returns the auth config for the given Docker image, extracting first its Docker registry.
// Finally, it will use the credential helpers to extract the information from the docker config file
// for that registry, if it exists.
func DockerImageAuth(ctx context.Context, image string) (string, registry.AuthConfig, error) {
	return auth.ForDockerImage(ctx, image)
}
