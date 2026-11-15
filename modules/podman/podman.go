// Package podman registers the Podman runtime for configuration-based selection.
// Import it and set TESTCONTAINERS_PROVIDER=podman or provider=podman in
// .testcontainers.properties. Importing the package does not connect to Podman.
package podman

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// Provider configures the Docker-compatible client for Podman's default network.
// Endpoint and socket discovery use the standard Testcontainers configuration.
type Provider struct{}

var _ testcontainers.RuntimeProvider = (*Provider)(nil)

func init() {
	if err := testcontainers.RegisterProvider("podman", func() testcontainers.RuntimeProvider {
		return &Provider{}
	}); err != nil {
		panic(err)
	}
}

func (*Provider) Initialize(ctx context.Context) (testcontainers.RuntimeConfig, error) {
	if err := ctx.Err(); err != nil {
		return testcontainers.RuntimeConfig{}, err
	}
	return testcontainers.RuntimeConfig{DefaultBridgeNetwork: "podman"}, nil
}
