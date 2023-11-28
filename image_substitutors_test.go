package testcontainers

import (
	"testing"

	"github.com/testcontainers/testcontainers-go/internal/config"
)

func TestPrependHubRegistrySubstitutor(t *testing.T) {
	resetConfigForTests := func() {
		config.Reset()
	}

	t.Run("should prepend the hub registry to images from Docker Hub", func(t *testing.T) {
		t.Run("plain image", func(t *testing.T) {
			t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "my-registry")
			defer resetConfigForTests()

			s := newPrependHubRegistry()

			img, err := s.Substitute("foo:latest")
			if err != nil {
				t.Fatal(err)
			}

			if img != "my-registry/foo:latest" {
				t.Errorf("expected my-registry/foo, got %s", img)
			}
		})
		t.Run("image with user", func(t *testing.T) {
			t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "my-registry")
			defer resetConfigForTests()

			s := newPrependHubRegistry()

			img, err := s.Substitute("user/foo:latest")
			if err != nil {
				t.Fatal(err)
			}

			if img != "my-registry/user/foo:latest" {
				t.Errorf("expected my-registry/foo, got %s", img)
			}
		})

		t.Run("image with organization and user", func(t *testing.T) {
			t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "my-registry")
			defer resetConfigForTests()

			s := newPrependHubRegistry()

			img, err := s.Substitute("org/user/foo:latest")
			if err != nil {
				t.Fatal(err)
			}

			if img != "my-registry/org/user/foo:latest" {
				t.Errorf("expected my-registry/org/foo:latest, got %s", img)
			}
		})
	})

	t.Run("should not prepend the hub registry to the image name", func(t *testing.T) {
		t.Run("non-hub image", func(t *testing.T) {
			t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "my-registry")
			defer resetConfigForTests()

			s := newPrependHubRegistry()

			img, err := s.Substitute("quay.io/foo:latest")
			if err != nil {
				t.Fatal(err)
			}

			if img != "quay.io/foo:latest" {
				t.Errorf("expected quay.io/foo:latest, got %s", img)
			}
		})

		t.Run("explicitly including docker.io", func(t *testing.T) {
			t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "my-registry")
			defer resetConfigForTests()

			s := newPrependHubRegistry()

			img, err := s.Substitute("docker.io/foo:latest")
			if err != nil {
				t.Fatal(err)
			}

			if img != "docker.io/foo:latest" {
				t.Errorf("expected docker.io/foo:latest, got %s", img)
			}
		})

		t.Run("explicitly including registry.hub.docker.com", func(t *testing.T) {
			t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "my-registry")
			defer resetConfigForTests()

			s := newPrependHubRegistry()

			img, err := s.Substitute("registry.hub.docker.com/foo:latest")
			if err != nil {
				t.Fatal(err)
			}

			if img != "registry.hub.docker.com/foo:latest" {
				t.Errorf("expected registry.hub.docker.com/foo:latest, got %s", img)
			}
		})
	})
}
