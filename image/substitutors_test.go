package image

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
)

func TestCustomHubSubstitutor(t *testing.T) {
	t.Run("prepend", func(t *testing.T) {
		s := NewCustomHubSubstitutor("quay.io")

		img, err := s.Substitute("foo/foo:latest")
		require.NoError(t, err)

		require.Equalf(t, "quay.io/foo/foo:latest", img, "expected quay.io/foo/foo:latest, got %s", img)
	})
	t.Run("not-prepend-as-prefix-exists", func(t *testing.T) {
		s := NewCustomHubSubstitutor("quay.io")

		img, err := s.Substitute("quay.io/foo/foo:latest")
		require.NoError(t, err)

		require.Equalf(t, "quay.io/foo/foo:latest", img, "expected quay.io/foo/foo:latest, got %s", img)
	})
	t.Run("not-prepend-as-config-hub-prefix-exists", func(t *testing.T) {
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
	t.Run("prepend", func(t *testing.T) {
		t.Run("plain-image", func(t *testing.T) {
			s := NewPrependHubRegistry("my-registry")

			img, err := s.Substitute("foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "my-registry/foo:latest", img, "expected my-registry/foo, got %s", img)
		})
		t.Run("image-with-user", func(t *testing.T) {
			s := NewPrependHubRegistry("my-registry")

			img, err := s.Substitute("user/foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "my-registry/user/foo:latest", img, "expected my-registry/foo, got %s", img)
		})

		t.Run("image-with-organization-and-user", func(t *testing.T) {
			s := NewPrependHubRegistry("my-registry")

			img, err := s.Substitute("org/user/foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "my-registry/org/user/foo:latest", img, "expected my-registry/org/foo:latest, got %s", img)
		})
	})

	t.Run("not-prepend", func(t *testing.T) {
		t.Run("non-hub-image", func(t *testing.T) {
			s := NewPrependHubRegistry("my-registry")

			img, err := s.Substitute("quay.io/foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "quay.io/foo:latest", img, "expected quay.io/foo:latest, got %s", img)
		})

		t.Run("prefix-is-docker-io", func(t *testing.T) {
			s := NewPrependHubRegistry("my-registry")

			img, err := s.Substitute("docker.io/foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "docker.io/foo:latest", img, "expected docker.io/foo:latest, got %s", img)
		})

		t.Run("prefix-is-registry-hub-docker-com", func(t *testing.T) {
			s := NewPrependHubRegistry("my-registry")

			img, err := s.Substitute("registry.hub.docker.com/foo:latest")
			require.NoError(t, err)

			require.Equalf(t, "registry.hub.docker.com/foo:latest", img, "expected registry.hub.docker.com/foo:latest, got %s", img)
		})
	})
}
