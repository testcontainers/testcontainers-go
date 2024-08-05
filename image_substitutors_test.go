package testcontainers

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go/image"
	"github.com/testcontainers/testcontainers-go/internal/config"
)

func TestSubstituteBuiltImage(t *testing.T) {
	req := Request{
		FromDockerfile: FromDockerfile{
			Context:    "testdata",
			Dockerfile: "echo.Dockerfile",
			Tag:        "my-image",
			Repo:       "my-repo",
		},
		ImageSubstitutors: []image.Substitutor{image.NewPrependHubRegistry("my-registry")},
		Started:           false,
	}

	t.Run("should not use the properties prefix on built images", func(t *testing.T) {
		config.Reset()
		c, err := Run(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}

		jsonRaw, err := c.Inspect(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if jsonRaw.Config.Image != "my-registry/my-repo:my-image" {
			t.Errorf("expected my-registry/my-repo:my-image, got %s", jsonRaw.Config.Image)
		}
	})
}
