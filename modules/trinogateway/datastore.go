package trinogateway

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
)

const (
	// aliasPostgres is the alias for the postgres container in the network.
	aliasPostgres = "postgres"

	// postgresDriver is the JDBC driver class for PostgreSQL.
	postgresDriver = "org.postgresql.Driver"
)

// DataStoreProvider provides the database used by Trino Gateway as data store, and owns the
// resources it creates. See ExternalProvider and ManagedPostgresProvider.
type DataStoreProvider interface {
	validate() error

	// start returns the connection settings of the database. If it fails, it removes
	// the resources it created.
	start(ctx context.Context, nw *testcontainers.DockerNetwork) (dataStoreConfig, error)

	terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error
}

var (
	_ DataStoreProvider = (*ExternalProvider)(nil)
	_ DataStoreProvider = (*ManagedPostgresProvider)(nil)
)

type dataStoreConfig struct {
	jdbcURL  string
	user     string
	password string
	driver   string
}

// ExternalProvider uses a database started by the caller, which must be reachable from the gateway.
// It creates no resources and never terminates the database.
type ExternalProvider struct {
	config dataStoreConfig
}

// NewExternalProvider returns a provider for the database at the given JDBC URL. The driver
// is the JDBC driver class, e.g. "org.postgresql.Driver".
func NewExternalProvider(jdbcURL, user, password, driver string) *ExternalProvider {
	return &ExternalProvider{
		config: dataStoreConfig{
			jdbcURL:  jdbcURL,
			user:     user,
			password: password,
			driver:   driver,
		},
	}
}

func (p *ExternalProvider) validate() error {
	switch {
	case p == nil:
		return errors.New("external provider must not be nil")
	case p.config.jdbcURL == "":
		return errors.New("external provider: JDBC URL must not be empty")
	case p.config.driver == "":
		return errors.New("external provider: driver must not be empty")
	}

	return nil
}

func (p *ExternalProvider) start(context.Context, *testcontainers.DockerNetwork) (dataStoreConfig, error) {
	return p.config, nil
}

// terminate is a no-op: the database is owned by the caller.
func (p *ExternalProvider) terminate(context.Context, ...testcontainers.TerminateOption) error {
	return nil
}

// ManagedPostgresProvider starts a PostgreSQL container on the gateway network,
// and terminates it with the gateway.
type ManagedPostgresProvider struct {
	img       string
	opts      []testcontainers.ContainerCustomizer
	container *postgres.PostgresContainer
}

// NewManagedPostgresProvider returns a provider that starts a PostgreSQL container with the
// given image and postgres module options.
func NewManagedPostgresProvider(img string, opts ...testcontainers.ContainerCustomizer) *ManagedPostgresProvider {
	return &ManagedPostgresProvider{img: img, opts: opts}
}

// Container returns the PostgreSQL container, or nil if it is not started or already terminated.
func (p *ManagedPostgresProvider) Container() *postgres.PostgresContainer {
	return p.container
}

func (p *ManagedPostgresProvider) validate() error {
	switch {
	case p == nil:
		return errors.New("managed postgres provider must not be nil")
	case p.img == "":
		return errors.New("managed postgres provider: image must not be empty")
	case p.container != nil:
		return errors.New("managed postgres provider: already started by another gateway")
	}

	return nil
}

func (p *ManagedPostgresProvider) start(ctx context.Context, nw *testcontainers.DockerNetwork) (dataStoreConfig, error) {
	opts := make([]testcontainers.ContainerCustomizer, 0, 2+len(p.opts))
	opts = append(opts,
		postgres.BasicWaitStrategies(),
		network.WithNetwork([]string{aliasPostgres}, nw),
	)
	opts = append(opts, p.opts...)

	ctr, err := postgres.Run(ctx, p.img, opts...)
	if err != nil {
		return dataStoreConfig{}, errors.Join(err, testcontainers.TerminateContainer(ctr))
	}

	cfg, err := postgresDataStore(ctx, ctr)
	if err != nil {
		return dataStoreConfig{}, errors.Join(err, testcontainers.TerminateContainer(ctr))
	}

	p.container = ctr
	return cfg, nil
}

func (p *ManagedPostgresProvider) terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	if p.container == nil {
		return nil
	}

	if err := p.container.Terminate(ctx, opts...); err != nil {
		return fmt.Errorf("terminate postgres: %w", err)
	}

	p.container = nil
	return nil
}

// postgresDataStore returns the connection settings of the postgres container, using its network alias.
func postgresDataStore(ctx context.Context, pg *postgres.PostgresContainer) (dataStoreConfig, error) {
	connStr, err := pg.ConnectionString(ctx)
	if err != nil {
		return dataStoreConfig{}, fmt.Errorf("postgres connection string: %w", err)
	}

	u, err := url.Parse(connStr)
	if err != nil {
		return dataStoreConfig{}, fmt.Errorf("parse postgres connection string: %w", err)
	}

	password, _ := u.User.Password()

	return dataStoreConfig{
		jdbcURL:  fmt.Sprintf("jdbc:postgresql://%s:5432%s", aliasPostgres, u.Path),
		user:     u.User.Username(),
		password: password,
		driver:   postgresDriver,
	}, nil
}
