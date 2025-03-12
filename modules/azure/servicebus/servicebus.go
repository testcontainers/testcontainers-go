package servicebus

import (
	"context"
	"errors"
	"fmt"

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

// Container represents the Azure Event Hubs container type used in the module
type Container struct {
	testcontainers.Container
	mssqlOptions *options
}

func (c *Container) MSSQLContainer() *mssql.MSSQLServerContainer {
	return c.mssqlOptions.mssqlContainer
}

// Terminate terminates the etcd container, its child nodes, and the network in which the cluster is running
// to communicate between the nodes.
func (c *Container) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	var errs []error

	if c.Container != nil {
		// terminate the eventhubs container
		if err := c.Container.Terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate eventhubs container: %w", err))
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

// Run creates an instance of the Azure Event Hubs container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image: img,
		Env: map[string]string{
			"SQL_WAIT_INTERVAL": "0", // default is zero because the MSSQL container is started first
		},
		ExposedPorts: []string{defaultPort, defaultHTTPPort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(defaultPort),
			wait.ForListeningPort(defaultHTTPPort),
			wait.ForHTTP("/health").WithPort(defaultHTTPPort),
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
			o(&defaultOptions)
		}
	}

	if genericContainerReq.Env["ACCEPT_EULA"] == "" {
		return nil, errors.New("EULA not accepted. Please use the WithAcceptEULA option to accept the EULA")
	}

	c := &Container{mssqlOptions: &defaultOptions}

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
			return nil, fmt.Errorf("run mssql container: %w", err)
		}
		defaultOptions.mssqlContainer = mssqlContainer

		genericContainerReq.Env["SQL_SERVER"] = aliasMSSQL
		genericContainerReq.Env["MSSQL_SA_PASSWORD"] = mssqlContainer.Password()

		// apply the network to the eventhubs container
		err = network.WithNetwork([]string{aliasServiceBus}, mssqlNetwork)(&genericContainerReq)
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
	hostPort, err := c.PortEndpoint(ctx, defaultPort, "")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return fmt.Sprintf(connectionStringFormat, hostPort, defaultSharedAccessKeyName, defaultSharedAccessKey), nil
}

// MustConnectionString returns the connection string for the eventhubs container,
// calling [Container.ConnectionString] and panicking if it returns an error.
func (c *Container) MustConnectionString(ctx context.Context) string {
	url, err := c.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	return url
}
