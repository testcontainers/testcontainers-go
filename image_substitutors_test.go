package testcontainers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/image"
	"github.com/testcontainers/testcontainers-go/internal/config"
)

// Deprecated: use testcontainers-go [image.NoopSubstitutor] instead
type NoopImageSubstitutor = image.NoopSubstitutor

// errorSubstitutor is a Substitutor that returns an error
type errorSubstitutor struct{}

var errSubstitution = errors.New("substitution error")

// Description returns a description of what is expected from this Substitutor,
// which is used in logs.
func (s errorSubstitutor) Description() string {
	return "errorSubstitutor"
}

// Substitute returns the original image, but returns an error
func (s errorSubstitutor) Substitute(image string) (string, error) {
	return image, errSubstitution
}

func TestImageSubstitutors(t *testing.T) {
	tests := []struct {
		name          string
		image         string // must be a valid image, as the test will try to create a container from it
		substitutors  []image.Substitutor
		expectedImage string
		expectedError error
	}{
		{
			name:          "no-substitutors",
			image:         "alpine",
			expectedImage: "alpine",
		},
		{
			name:          "noop-substitutor",
			image:         "alpine",
			substitutors:  []image.Substitutor{image.NoopSubstitutor{}},
			expectedImage: "alpine",
		},
		{
			name:          "prepend-namespace",
			image:         "alpine",
			substitutors:  []image.Substitutor{image.DockerSubstitutor{}},
			expectedImage: "registry.hub.docker.com/library/alpine",
		},
		{
			name:          "substitution-with-error",
			image:         "alpine",
			substitutors:  []image.Substitutor{errorSubstitutor{}},
			expectedImage: "alpine",
			expectedError: errSubstitution,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			req := ContainerRequest{
				Image:             test.image,
				ImageSubstitutors: test.substitutors,
			}

			ctr, err := GenericContainer(ctx, GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			CleanupContainer(t, ctr)
			if test.expectedError != nil {
				require.ErrorIs(t, err, test.expectedError)
				return
			}

			require.NoError(t, err)

			// enforce the concrete type, as GenericContainer returns an interface,
			// which will be changed in future implementations of the library
			dockerContainer := ctr.(*DockerContainer)
			require.Equal(t, test.expectedImage, dockerContainer.Image)
		})
	}
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
			ImageSubstitutors: []image.Substitutor{image.NewPrependHubRegistry("my-registry")},
		},
		Started: false,
	}

	t.Run("should-use-image-substitutors", func(t *testing.T) {
		config.Reset()
		c, err := GenericContainer(context.Background(), req)
		CleanupContainer(t, c)
		require.NoError(t, err)

		json, err := c.Inspect(context.Background())
		require.NoError(t, err)

		require.Equalf(t, "my-registry/my-repo:my-image", json.Config.Image, "expected my-registry/my-repo:my-image, got %s", json.Config.Image)
	})
}
