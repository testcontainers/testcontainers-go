package testcontainers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
)

func TestCustomHubSubstitutor(t *testing.T) {
	t.Run("should substitute the image with the provided one", func(t *testing.T) {
		s := NewCustomHubSubstitutor("quay.io")

		img, err := s.Substitute("foo/foo:latest")
		require.NoError(t, err)

		require.Equalf(t, "quay.io/foo/foo:latest", img, "expected quay.io/foo/foo:latest, got %s", img)
	})
	t.Run("should not substitute the image if it is already using the provided hub", func(t *testing.T) {
		s := NewCustomHubSubstitutor("quay.io")

		img, err := s.Substitute("quay.io/foo/foo:latest")
		require.NoError(t, err)

		require.Equalf(t, "quay.io/foo/foo:latest", img, "expected quay.io/foo/foo:latest, got %s", img)
	})
	t.Run("should not substitute the image if hub image name prefix config exist", func(t *testing.T) {
		t.Cleanup(config.Reset)
		config.Reset()
		t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "registry.mycompany.com/mirror")
		s := NewCustomHubSubstitutor("quay.io")

		img, err := s.Substitute("foo/foo:latest")
		require.NoError(t, err)

		require.Equalf(t, "foo/foo:latest", img, "expected foo/foo:latest, got %s", img)
	})
}

func TestPrependHubRegistrySubstitutor(t *testing.T) {
	t.Run("should prepend the hub registry to images from Docker Hub", func(t *testing.T) {
		t.Run("plain image", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

			img, err := s.Substitute("foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "my-registry/foo:latest", img, "expected my-registry/foo, got %s", img)
		})
		t.Run("image with user", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

			img, err := s.Substitute("user/foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "my-registry/user/foo:latest", img, "expected my-registry/foo, got %s", img)
		})

		t.Run("image with organization and user", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

			img, err := s.Substitute("org/user/foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "my-registry/org/user/foo:latest", img, "expected my-registry/org/foo:latest, got %s", img)
		})
	})

	t.Run("should not prepend the hub registry to the image name", func(t *testing.T) {
		t.Run("non-hub image", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

			img, err := s.Substitute("quay.io/foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "quay.io/foo:latest", img, "expected quay.io/foo:latest, got %s", img)
		})

		t.Run("explicitly including registry.hub.docker.com/library", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

			img, err := s.Substitute("registry.hub.docker.com/library/foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "registry.hub.docker.com/library/foo:latest", img, "expected registry.hub.docker.com/library/foo:latest, got %s", img)
		})

		t.Run("explicitly including registry.hub.docker.com", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

			img, err := s.Substitute("registry.hub.docker.com/foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "registry.hub.docker.com/foo:latest", img, "expected registry.hub.docker.com/foo:latest, got %s", img)
		})
	})
}

func TestSubstituteBuiltImage(t *testing.T) {
	req := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			FromDockerfile: FromDockerfile{
				Context:    "testdata",
				Dockerfile: "echo.Dockerfile",
				Tag:        "my-image",
				Repo:       "my-repo",
			},
			ImageSubstitutors: []ImageSubstitutor{newPrependHubRegistry("my-registry")},
		},
		Started: false,
	}

	t.Run("should not use the properties prefix on built images", func(t *testing.T) {
		config.Reset()
		c, err := GenericContainer(context.Background(), req)
		CleanupContainer(t, c)
		require.NoError(t, err)

		json, err := c.Inspect(context.Background())
		require.NoError(t, err)

		require.Equalf(t, "my-registry/my-repo:my-image", json.Config.Image, "expected my-registry/my-repo:my-image, got %s", json.Config.Image)
	})
}
