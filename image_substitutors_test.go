package testcontainers

import (
	"testing"

	"github.com/testcontainers/testcontainers-go/internal/config"
)

func TestPrependHubRegistrySubstitutor(t *testing.T) {
	t.Run("should prepend the hub registry to the image name", func(t *testing.T) {
		t.Setenv("TESTCONTAINERS_HUB_IMAGE_NAME_PREFIX", "my-registry")
		defer config.Reset()

		s := newPrependHubRegistry()

		img, err := s.Substitute("foo")
		if err != nil {
			t.Fatal(err)
		}

		if img != "my-registry/foo" {
			t.Errorf("expected my-registry/foo, got %s", img)
		}
	})
}
