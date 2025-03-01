package image

import (
	"fmt"
	"net/url"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

// imageSubstitutor {

// Substitutor represents a way to substitute container image names
type Substitutor interface {
	// Description returns the name of the type and a short description of how it modifies the image.
	// Useful to be printed in logs
	Description() string
	Substitute(image string) (string, error)
}

// }

// CustomHubSubstitutor represents a way to substitute the hub of an image with a custom one,
// using provided value with respect to the HubImageNamePrefix configuration value.
type CustomHubSubstitutor struct {
	hub string
}

// NewCustomHubSubstitutor creates a new CustomHubSubstitutor
func NewCustomHubSubstitutor(hub string) CustomHubSubstitutor {
	return CustomHubSubstitutor{
		hub: hub,
	}
}

// Description returns the name of the type and a short description of how it modifies the image.
func (c CustomHubSubstitutor) Description() string {
	return fmt.Sprintf("CustomHubSubstitutor (replaces hub with %s)", c.hub)
}

// Substitute replaces the hub of the image with the provided one, with certain conditions:
//   - if the hub is empty, the image is returned as is.
//   - if the image already contains a registry, the image is returned as is.
//   - if the HubImageNamePrefix configuration value is set in the Testcontainers configuration, the image is returned as is.
func (c CustomHubSubstitutor) Substitute(image string) (string, error) {
	registry := core.ExtractRegistry(image, "")
	cfg := config.Read()

	exclusions := []func() bool{
		func() bool { return c.hub == "" },
		func() bool { return registry != "" },
		func() bool { return cfg.HubImageNamePrefix != "" },
	}

	for _, exclusion := range exclusions {
		if exclusion() {
			return image, nil
		}
	}

	result, err := url.JoinPath(c.hub, image)
	if err != nil {
		return "", err
	}

	return result, nil
}

// PrependHubRegistry represents a way to prepend a custom Hub registry to the image name,
// using the HubImageNamePrefix configuration value
type PrependHubRegistry struct {
	prefix string
}

// NewPrependHubRegistry creates a new PrependHubRegistry
func NewPrependHubRegistry(hubPrefix string) PrependHubRegistry {
	return PrependHubRegistry{
		prefix: hubPrefix,
	}
}

// Description returns the name of the type and a short description of how it modifies the image.
func (p PrependHubRegistry) Description() string {
	return fmt.Sprintf("HubImageSubstitutor (prepends %s)", p.prefix)
}

// Substitute prepends the Hub prefix to the image name, with certain conditions:
//   - if the prefix is empty, the image is returned as is.
//   - if the image is a non-hub image (e.g. where another registry is set), the image is returned as is.
//   - if the image is a Docker Hub image where the hub registry is explicitly part of the name
//     (i.e. anything with a registry.hub.docker.com host part), the image is returned as is.
func (p PrependHubRegistry) Substitute(image string) (string, error) {
	registry := core.ExtractRegistry(image, "")

	// add the exclusions in the right order
	exclusions := []func() bool{
		func() bool { return p.prefix == "" },                        // no prefix set at the configuration level
		func() bool { return registry != "" },                        // non-hub image
		func() bool { return registry == "docker.io" },               // explicitly including docker.io
		func() bool { return registry == "registry.hub.docker.com" }, // explicitly including registry.hub.docker.com
	}

	for _, exclusion := range exclusions {
		if exclusion() {
			return image, nil
		}
	}

	result, err := url.JoinPath(p.prefix, image)
	if err != nil {
		return "", err
	}

	return result, nil
}

// noopImageSubstitutor {
// NoopSubstitutor is a Substitutor that returns the original image, without any change.
type NoopSubstitutor struct{}

// Description returns a description of what is expected from this Substitutor,
// which is used in logs.
func (s NoopSubstitutor) Description() string {
	return "NoopSubstitutor (noop)"
}

// Substitute returns the original image, without any change
func (s NoopSubstitutor) Substitute(image string) (string, error) {
	return image, nil
}

// }

// dockerImageSubstitutor {
// DockerSubstitutor is a Substitutor that prepends the registry.hub.docker.com prefix to the image name.
type DockerSubstitutor struct{}

// Description returns a description of what is expected from this Substitutor,
// which is used in logs.
func (s DockerSubstitutor) Description() string {
	return "DockerSubstitutor (prepends registry.hub.docker.com)"
}

// Substitute prepends the registry.hub.docker.com prefix to the image name.
func (s DockerSubstitutor) Substitute(image string) (string, error) {
	return "registry.hub.docker.com/library/" + image, nil
}

// }
