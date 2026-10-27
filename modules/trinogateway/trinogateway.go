package trinogateway

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// httpPort is the default port exposed by Trino Gateway.
	httpPort = "8080/tcp"

	// startupTimeout is the time to wait for Trino Gateway to be ready, including the database migrations.
	startupTimeout = 3 * time.Minute

	// configPath is the path of the configuration file inside the Trino Gateway container.
	configPath = "/etc/trino-gateway/config.yaml"

	// aliasGateway is the network alias of the Trino Gateway container.
	aliasGateway = "trino-gateway"

	// aliasPostgres is the network alias of the PostgreSQL container used as data store.
	aliasPostgres = "postgres"

	// defaultPostgresImage is the default image of the PostgreSQL container used as data store.
	defaultPostgresImage = "postgres:17-alpine"

	// postgresDriver is the JDBC driver class for PostgreSQL.
	postgresDriver = "org.postgresql.Driver"

	// Environment variables referenced by the default configuration file, see config.yaml.
	// A custom configuration file passed with WithConfig can reference them too, using
	// the "${ENV:NAME}" syntax supported by Trino Gateway.
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
	postgres *postgres.PostgresContainer
	network  *testcontainers.DockerNetwork
}

// ConnectionString returns the HTTP connection string for the Trino Gateway,
// e.g. "http://localhost:8080". Trino clients connect to this URL as if it
// were a Trino coordinator.
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, httpPort, "http")
}

// Network returns the network the Trino Gateway container is attached to.
// Trino backends must be attached to this network so the gateway can reach them.
func (c *Container) Network() *testcontainers.DockerNetwork {
	return c.network
}

// PostgresContainer returns the PostgreSQL container used as data store,
// or nil if an external data store was configured with WithDataStore.
func (c *Container) PostgresContainer() *postgres.PostgresContainer {
	return c.postgres
}

// Terminate terminates the Trino Gateway container, the PostgreSQL container, and the network to communicate between them.
func (c *Container) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	var errs []error

	if c.Container != nil {
		if err := c.Container.Terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate trino gateway container: %w", err))
		}
	}

	if c.postgres != nil {
		if err := c.postgres.Terminate(ctx, opts...); err != nil {
			errs = append(errs, fmt.Errorf("terminate postgres container: %w", err))
		}
	}

	if c.network != nil {
		if err := c.network.Remove(ctx); err != nil {
			errs = append(errs, fmt.Errorf("remove network: %w", err))
		}
	}

	return errors.Join(errs...)
}

// Run creates an instance of the Trino Gateway container type.
//
// Trino Gateway requires a relational database as data store. By default, a PostgreSQL
// container is started and attached, together with the gateway, to a dedicated network.
// Use WithPostgres to customise it, or WithDataStore to use an external database instead.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	settings := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&settings); err != nil {
				return nil, fmt.Errorf("trinogateway option: %w", err)
			}
		}
	}

	nw, err := network.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("new network: %w", err)
	}

	// From here on, the container is always returned so the caller can terminate
	// the resources that were created, even when an error occurs.
	c := &Container{network: nw}

	ds := settings.dataStore
	if ds == nil {
		pgOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(settings.postgresOptions))
		pgOpts = append(pgOpts,
			postgres.BasicWaitStrategies(),
			network.WithNetwork([]string{aliasPostgres}, nw),
		)
		pgOpts = append(pgOpts, settings.postgresOptions...)

		// Start the postgres container first: the gateway runs the database migrations on startup.
		pg, err := postgres.Run(ctx, settings.postgresImage, pgOpts...)
		c.postgres = pg
		if err != nil {
			return c, fmt.Errorf("run postgres: %w", err)
		}

		ds, err = postgresDataStore(ctx, pg)
		if err != nil {
			return c, err
		}
	}

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

// postgresDataStore builds the data store settings for the postgres container,
// reading the credentials and database name from its connection string and
// pointing the JDBC URL to the network alias of the container.
func postgresDataStore(ctx context.Context, pg *postgres.PostgresContainer) (*dataStore, error) {
	connStr, err := pg.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres connection string: %w", err)
	}

	u, err := url.Parse(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse postgres connection string: %w", err)
	}

	password, _ := u.User.Password()

	return &dataStore{
		jdbcURL:  fmt.Sprintf("jdbc:postgresql://%s:5432%s", aliasPostgres, u.Path),
		user:     u.User.Username(),
		password: password,
		driver:   postgresDriver,
	}, nil
}
