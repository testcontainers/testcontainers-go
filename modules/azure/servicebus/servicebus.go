package servicebus

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort                = "5672/tcp"
	defaultHTTPPort            = "5300/tcp"
	defaultSharedAccessKeyName = "RootManageSharedAccessKey"
	defaultSharedAccessKey     = "SAS_KEY_VALUE"
	connectionStringFormat     = "Endpoint=sb://%s;SharedAccessKeyName=%s;SharedAccessKey=%s;UseDevelopmentEmulator=true;"

	// aliasServiceBus is the alias for the servicebus container in the network
	aliasServiceBus = "servicebus"

	// aliasMSSQL is the alias for the mssql network
	aliasMSSQL = "mssql"

	// defaultMSSQLImage is the default image for the mssql container
	defaultMSSQLImage = "mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04"

	// containerConfigFile is the path to the config file for the servicebus container
	containerConfigFile = "/ServiceBus_Emulator/ConfigFiles/Config.json"
)

// Container represents the Azure ServiceBus container type used in the module
type Container struct {
	testcontainers.Container
	mssqlOptions *options
}

// MSSQLContainer returns the mssql container that is used by the servicebus container
func (c *Container) MSSQLContainer() *mssql.MSSQLServerContainer {
	return c.mssqlOptions.mssqlContainer
}

// Terminate terminates the servicebus container, the mssql container, and the network to communicate between them.
func (c *Container) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	var errs []error

	if c.Container != nil {
		// terminate the servicebus container
		if err := c.Container.Terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate servicebus container: %w", err))
		}
	}

	// terminate the mssql container if it was created
	if c.mssqlOptions.mssqlContainer != nil {
		if err := c.mssqlOptions.mssqlContainer.Terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate mssql container: %w", err))
		}
	}

	// remove the mssql network if it was created
	if c.mssqlOptions.network != nil {
		if err := c.mssqlOptions.network.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("remove mssql network: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Run creates an instance of the Azure ServiceBus container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// Process custom options first to extract settings
	defaultOptions := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&defaultOptions); err != nil {
				return nil, fmt.Errorf("servicebus option: %w", err)
			}
		}
	}

	c := &Container{mssqlOptions: &defaultOptions}

	// Build moduleOpts with defaults
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultPort, defaultHTTPPort),
		testcontainers.WithEnv(map[string]string{
			"SQL_WAIT_INTERVAL": "0", // default is zero because the MSSQL container is started first
		}),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort(defaultPort),
			wait.ForListeningPort(defaultHTTPPort),
			wait.ForHTTP("/health").WithPort(defaultHTTPPort).WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
		),
	}

	if defaultOptions.mssqlContainer == nil {
		mssqlNetwork, err := network.New(ctx)
		if err != nil {
			return c, fmt.Errorf("new mssql network: %w", err)
		}
		defaultOptions.network = mssqlNetwork

		mssqlOpts := []testcontainers.ContainerCustomizer{
			mssql.WithAcceptEULA(),
			network.WithNetwork([]string{aliasMSSQL}, mssqlNetwork),
		}

		mssqlOpts = append(mssqlOpts, defaultOptions.mssqlOptions...)

		// Start the mssql container first. The EULA is accepted by default, as it is required by the servicebus emulator.
		mssqlContainer, err := mssql.Run(ctx, defaultOptions.mssqlImage, mssqlOpts...)
		if err != nil {
			return c, fmt.Errorf("run mssql container: %w", err)
		}
		defaultOptions.mssqlContainer = mssqlContainer

		moduleOpts = append(moduleOpts, testcontainers.WithEnv(map[string]string{
			"SQL_SERVER":        aliasMSSQL,
			"MSSQL_SA_PASSWORD": mssqlContainer.Password(),
		}))

		// apply the network to the servicebus container
		moduleOpts = append(moduleOpts, network.WithNetwork([]string{aliasServiceBus}, mssqlNetwork))
	}

	moduleOpts = append(moduleOpts, opts...)

	// validate the EULA after all the options are applied
	moduleOpts = append(moduleOpts, validateEula())

	var err error
	c.Container, err = testcontainers.Run(ctx, img, moduleOpts...)
	if err != nil {
		return c, fmt.Errorf("run servicebus: %w", err)
	}

	return c, nil
}

// ConnectionString returns the connection string for the servicebus container,
// using the following format:
// Endpoint=sb://<hostname>:<port>;SharedAccessKeyName=<key-name>;SharedAccessKey=<key>;UseDevelopmentEmulator=true;
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	// we are passing an empty proto to get the host:port string
	hostPort, err := c.PortEndpoint(ctx, defaultPort, "")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return fmt.Sprintf(connectionStringFormat, hostPort, defaultSharedAccessKeyName, defaultSharedAccessKey), nil
}
