package testcontainers

import (
	"context"
	"testing"
)

func TestPrependHubRegistrySubstitutor(t *testing.T) {
	t.Run("should prepend the hub registry to images from Docker Hub", func(t *testing.T) {
		t.Run("plain image", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

			img, err := s.Substitute("foo:latest")
			if err != nil {
				t.Fatal(err)
			}

			if img != "my-registry/foo:latest" {
				t.Errorf("expected my-registry/foo, got %s", img)
			}
		})
		t.Run("image with user", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

			img, err := s.Substitute("user/foo:latest")
			if err != nil {
				t.Fatal(err)
			}

			if img != "my-registry/user/foo:latest" {
				t.Errorf("expected my-registry/foo, got %s", img)
			}
		})

		t.Run("image with organization and user", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

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
			s := newPrependHubRegistry("my-registry")

			img, err := s.Substitute("quay.io/foo:latest")
			if err != nil {
				t.Fatal(err)
			}

			if img != "quay.io/foo:latest" {
				t.Errorf("expected quay.io/foo:latest, got %s", img)
			}
		})

		t.Run("explicitly including docker.io", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

			img, err := s.Substitute("docker.io/foo:latest")
			if err != nil {
				t.Fatal(err)
			}

			if img != "docker.io/foo:latest" {
				t.Errorf("expected docker.io/foo:latest, got %s", img)
			}
		})

		t.Run("explicitly including registry.hub.docker.com", func(t *testing.T) {
			s := newPrependHubRegistry("my-registry")

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
		c, err := GenericContainer(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}

		json, err := c.Inspect(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if json.Config.Image != "my-registry/my-repo:my-image" {
			t.Errorf("expected my-registry/my-repo:my-image, got %s", json.Config.Image)
		}
	})
}
