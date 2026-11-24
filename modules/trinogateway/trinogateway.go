package trinogateway

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// httpPort is the default port exposed by Trino Gateway.
	httpPort = "8080/tcp"

	// startupTimeout includes the time to run the database migrations.
	startupTimeout = 3 * time.Minute

	// configPath is the path of the config file in the container.
	configPath = "/etc/trino-gateway/config.yaml"

	// aliasGateway is the alias for the gateway container in the network.
	aliasGateway = "trino-gateway"

	// Environment variables with the data store settings, referenced by config.yaml.
	envDataStoreJDBCURL  = "DATASTORE_JDBC_URL"
	envDataStoreUser     = "DATASTORE_USER"
	envDataStorePassword = "DATASTORE_PASSWORD"
	envDataStoreDriver   = "DATASTORE_DRIVER"
)

//go:embed config.yaml
var defaultConfig string

// Container represents the Trino Gateway container type used in the module.
type Container struct {
	testcontainers.Container
	dataStore DataStoreProvider
	network   *testcontainers.DockerNetwork
}

// ConnectionString returns the HTTP connection string for the Trino Gateway,
// e.g. "http://localhost:8080".
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, httpPort, "http")
}

// Network returns the network created for the gateway. Trino backends must be attached to it.
func (c *Container) Network() *testcontainers.DockerNetwork {
	return c.network
}

// Terminate terminates the gateway container, the resources created by the data store
// provider, and the network, in this order.
func (c *Container) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	var errs []error

	if c.Container != nil {
		if err := c.Container.Terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate trino gateway container: %w", err))
		}
	}

	if c.dataStore != nil {
		if err := c.dataStore.terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate data store: %w", err))
		}
	}

	if c.network != nil {
		if err := c.network.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("remove network: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Run creates an instance of the Trino Gateway container type. The data store must be
// selected with WithDataStore.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	var settings options
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&settings); err != nil {
				return nil, fmt.Errorf("trinogateway option: %w", err)
			}
		}
	}

	if settings.dataStore == nil {
		return nil, errors.New("trinogateway: a data store provider is required, use WithDataStore")
	}

	nw, err := network.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("new network: %w", err)
	}

	// From here on, return the container so the caller can terminate what was created.
	c := &Container{network: nw}

	// Start the data store first: the gateway runs the database migrations on startup.
	ds, err := settings.dataStore.start(ctx, nw)
	if err != nil {
		return c, fmt.Errorf("start data store: %w", err)
	}
	c.dataStore = settings.dataStore

	var config io.Reader = strings.NewReader(defaultConfig)
	if settings.config != nil {
		config = settings.config
	}

	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 5+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(httpPort),
		network.WithNetwork([]string{aliasGateway}, nw),
		testcontainers.WithEnv(map[string]string{
			envDataStoreJDBCURL:  ds.jdbcURL,
			envDataStoreUser:     ds.user,
			envDataStorePassword: ds.password,
			envDataStoreDriver:   ds.driver,
		}),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader:            config,
			ContainerFilePath: configPath,
			FileMode:          0o644,
		}),
		testcontainers.WithWaitStrategyAndDeadline(startupTimeout,
			// the gateway reports ready once the cluster monitor has completed its first run
			wait.ForHTTP("/trino-gateway/readyz").WithPort(httpPort).WithStartupTimeout(startupTimeout),
			// listing the backends verifies that the data store is reachable and migrated
			wait.ForHTTP("/api/public/backends").WithPort(httpPort).WithStartupTimeout(startupTimeout),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	if ctr != nil {
		c.Container = ctr
	}

	if err != nil {
		return c, fmt.Errorf("run trinogateway: %w", err)
	}

	return c, nil
}
