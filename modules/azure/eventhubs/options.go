package eventhubs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/azurite"
)

type options struct {
	azuriteImage     string
	azuriteOptions   []testcontainers.ContainerCustomizer
	azuriteContainer *azurite.Container
	network          *testcontainers.DockerNetwork

	// azuriteAlias is the network alias of the azurite container; defaults to aliasAzurite ("azurite").
	azuriteAlias string
	// azuriteOwned is true when the eventhubs module owns the azurite container lifecycle.
	// When false (user-supplied via WithAzuriteContainer), Terminate() will not tear down azurite or network.
	azuriteOwned bool
	// azuriteUserProvided is true when WithAzuriteContainer has been applied.
	// Used for mutual-exclusion detection with WithAzurite.
	azuriteUserProvided bool
	// azuriteModuleOwned is true when WithAzurite was explicitly called.
	// Used for mutual-exclusion detection with WithAzuriteContainer.
	azuriteModuleOwned bool
}

func defaultOptions() options {
	return options{
		azuriteImage: "mcr.microsoft.com/azure-storage/azurite:3.33.0",
		azuriteOwned: true,
	}
}

// Satisfy the testcontainers.ContainerCustomizer interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the EventHubs container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithAzurite sets the image and options for the Azurite container that the
// module will create automatically.
// By default, the image is "mcr.microsoft.com/azure-storage/azurite:3.33.0".
// It is mutually exclusive with WithAzuriteContainer: passing both options to
// Run returns an error. If you already have a running Azurite container, use
// WithAzuriteContainer instead — the image and options set here will not apply.
func WithAzurite(img string, opts ...testcontainers.ContainerCustomizer) Option {
	return func(o *options) error {
		if o.azuriteUserProvided {
			return errors.New("eventhubs: WithAzurite and WithAzuriteContainer are mutually exclusive")
		}
		o.azuriteImage = img
		o.azuriteOptions = opts
		o.azuriteModuleOwned = true
		return nil
	}
}

// WithAzuriteContainer wires an already-running Azurite container into the
// Event Hubs emulator. The caller owns the lifecycle of both the network and
// the Azurite container: the eventhubs container will NOT terminate them on
// Terminate(). This is the key behavioural difference vs. the default path
// where eventhubs owns Azurite.
//
// The network must already contain the Azurite container under the provided
// alias (default aliasAzurite when alias is empty). The Event Hubs container
// is attached to the same network under the aliasEventhubs alias.
//
// It returns an error if azuriteContainer or dockerNetwork is nil, or if
// WithAzurite has already been applied.
func WithAzuriteContainer(
	azuriteContainer *azurite.Container,
	dockerNetwork *testcontainers.DockerNetwork,
	alias string,
) Option {
	return func(o *options) error {
		if azuriteContainer == nil {
			return errors.New("eventhubs: azurite container is nil")
		}
		if dockerNetwork == nil {
			return errors.New("eventhubs: docker network is nil")
		}
		if o.azuriteModuleOwned {
			return errors.New("eventhubs: WithAzuriteContainer and WithAzurite are mutually exclusive")
		}
		if o.azuriteUserProvided {
			return errors.New("eventhubs: WithAzuriteContainer called more than once")
		}
		if alias == "" {
			alias = aliasAzurite
		}
		o.azuriteContainer = azuriteContainer
		o.network = dockerNetwork
		o.azuriteAlias = alias
		o.azuriteOwned = false
		o.azuriteUserProvided = true
		return nil
	}
}

// WithAcceptEULA sets the ACCEPT_EULA environment variable to "Y" for the eventhubs container.
func WithAcceptEULA() testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"ACCEPT_EULA": "Y",
	})
}

// WithConfig sets the eventhubs config file for the eventhubs container,
// copying the content of the reader to the container file at
// "/Eventhubs_Emulator/ConfigFiles/Config.json".
func WithConfig(r io.Reader) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Files = append(req.Files, testcontainers.ContainerFile{
			Reader:            r,
			ContainerFilePath: containerConfigFile,
			FileMode:          0o644,
		})

		return nil
	}
}

// WithConfigObject marshals the supplied *Config to JSON and injects it at
// containerConfigFile (/Eventhubs_Emulator/ConfigFiles/Config.json).
// This is the statically-typed counterpart to WithConfig.
// Returns an error if cfg is nil or if JSON marshalling fails.
func WithConfigObject(cfg *Config) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if cfg == nil {
			return errors.New("eventhubs: config is nil")
		}
		b, err := json.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("eventhubs: marshal config: %w", err)
		}
		req.Files = append(req.Files, testcontainers.ContainerFile{
			Reader:            bytes.NewReader(b),
			ContainerFilePath: containerConfigFile,
			FileMode:          0o644,
		})
		return nil
	}
}

// validateEula validates that the EULA is accepted for the eventhubs container.
func validateEula() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if strings.ToUpper(req.Env["ACCEPT_EULA"]) != "Y" {
			return errors.New("EULA not accepted. Please use the WithAcceptEULA option to accept the EULA")
		}

		return nil
	}
}
