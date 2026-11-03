package testcontainers

import "github.com/testcontainers/testcontainers-go/internal/core"

// RuntimeProvider prepares a Docker-compatible runtime before any Docker host,
// socket, or client is discovered. Its Initialize method runs once per process.
// Implementations must not call testcontainers constructors from Initialize.
// The initialization context preserves the first caller's values but not its
// cancellation or deadline. Implementations should set their own startup timeout.
// Canceling an individual caller stops its wait without canceling shared startup.
type RuntimeProvider = core.RuntimeProvider

// RuntimeConfig supplies the client endpoint, daemon-side socket path, published
// port host, and default bridge network. Empty fields use standard discovery.
// Name is assigned by the registry, regardless of the value returned by a provider.
type RuntimeConfig = core.RuntimeConfig

// ProviderFactory constructs a provider without starting its runtime on import.
type ProviderFactory = core.ProviderFactory

// RegisterProvider registers a named runtime, normally from a module's init
// function. Registration must precede Docker discovery. Names are case insensitive;
// duplicate names and attempts to replace the built-in Docker provider are errors.
// Select a registered provider through TESTCONTAINERS_PROVIDER or the provider
// property. A module must be imported; adding it to go.mod alone does not register it.
func RegisterProvider(name string, factory ProviderFactory) error {
	return core.RegisterProvider(name, factory)
}
