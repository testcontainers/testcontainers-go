package eventhubs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/azurite"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultAMPQPort        = "5672/tcp"
	defaultHTTPPort        = "5300/tcp"
	connectionStringFormat = "Endpoint=sb://%s;SharedAccessKeyName=%s;SharedAccessKey=%s;UseDevelopmentEmulator=true;"

	// aliasEventhubs is the alias for the eventhubs container in the network
	aliasEventhubs = "eventhubs"

	// aliasAzurite is the alias for the azurite container in the network
	aliasAzurite = "azurite"

	// containerConfigFile is the path to the eventhubs config file
	containerConfigFile = "/Eventhubs_Emulator/ConfigFiles/Config.json"
)

// Container represents the Azure Event Hubs container type used in the module
type Container struct {
	testcontainers.Container
	azuriteOptions *options
}

// AzuriteContainer returns the azurite container that is used by the eventhubs container
func (c *Container) AzuriteContainer() *azurite.Container {
	return c.azuriteOptions.azuriteContainer
}

// Terminate terminates the eventhubs container, the azurite container, and the network to communicate between them.
func (c *Container) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	var errs []error

	if c.Container != nil {
		// terminate the eventhubs container
		if err := c.Container.Terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate eventhubs container: %w", err))
		}
	}

	// terminate the azurite container if it was created
	if c.azuriteOptions.azuriteContainer != nil {
		if err := c.azuriteOptions.azuriteContainer.Terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate azurite container: %w", err))
		}
	}

	// remove the azurite network if it was created
	if c.azuriteOptions.network != nil {
		if err := c.azuriteOptions.network.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("remove azurite network: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Run creates an instance of the Azure Event Hubs container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{defaultAMPQPort, defaultHTTPPort},
		Env:          make(map[string]string),
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(defaultAMPQPort),
			wait.ForListeningPort(defaultHTTPPort),
			wait.ForHTTP("/health").WithPort(defaultHTTPPort).WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
		),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	defaultOptions := defaultOptions()
	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
		if o, ok := opt.(Option); ok {
			if err := o(&defaultOptions); err != nil {
				return nil, fmt.Errorf("eventhubsoption: %w", err)
			}
		}
	}

	if strings.ToUpper(genericContainerReq.Env["ACCEPT_EULA"]) != "Y" {
		return nil, errors.New("EULA not accepted. Please use the WithAcceptEULA option to accept the EULA")
	}

	c := &Container{azuriteOptions: &defaultOptions}

	if defaultOptions.azuriteContainer == nil {
		azuriteNetwork, err := network.New(ctx)
		if err != nil {
			return c, fmt.Errorf("new azurite network: %w", err)
		}
		defaultOptions.network = azuriteNetwork

		azuriteOpts := []testcontainers.ContainerCustomizer{
			network.WithNetwork([]string{aliasAzurite}, azuriteNetwork),
		}
		azuriteOpts = append(azuriteOpts, defaultOptions.azuriteOptions...)

		// start the azurite container first
		azuriteContainer, err := azurite.Run(ctx, defaultOptions.azuriteImage, azuriteOpts...)
		if err != nil {
			return c, fmt.Errorf("run azurite container: %w", err)
		}
		defaultOptions.azuriteContainer = azuriteContainer

		genericContainerReq.Env["BLOB_SERVER"] = aliasAzurite
		genericContainerReq.Env["METADATA_SERVER"] = aliasAzurite

		// apply the network to the eventhubs container
		err = network.WithNetwork([]string{aliasEventhubs}, azuriteNetwork)(&genericContainerReq)
		if err != nil {
			return c, fmt.Errorf("with network: %w", err)
		}
	}

	var err error
	c.Container, err = testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// ConnectionString returns the connection string for the eventhubs container,
// using the following format:
// Endpoint=sb://<hostname>:<port>;SharedAccessKeyName=<key-name>;SharedAccessKey=<key>;UseDevelopmentEmulator=true;
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	// we are passing an empty proto to get the host:port string
	hostPort, err := c.PortEndpoint(ctx, defaultAMPQPort, "")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return fmt.Sprintf(connectionStringFormat, hostPort, azurite.AccountName, azurite.AccountKey), nil
}
