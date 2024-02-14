package postgres

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

const (
	defaultUser          = "postgres"
	defaultPassword      = "postgres"
	defaultPostgresImage = "docker.io/postgres:11-alpine"
	defaultSnapshotName  = "migrated_template"
)

// PostgresContainer represents the postgres container type used in the module
type PostgresContainer struct {
	testcontainers.Container
	dbName       string
	user         string
	password     string
	snapshotName string
}

// ConnectionString returns the connection string for the postgres container, using the default 5432 port, and
// obtaining the host and exposed port from the container. It also accepts a variadic list of extra arguments
// which will be appended to the connection string. The format of the extra arguments is the same as the
// connection string format, e.g. "connect_timeout=10" or "application_name=myapp"
func (c *PostgresContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	containerPort, err := c.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	extraArgs := strings.Join(args, "&")
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?%s", c.user, c.password, net.JoinHostPort(host, containerPort.Port()), c.dbName, extraArgs)
	return connStr, nil
}

// WithConfigFile sets the config file to be used for the postgres container
// It will also set the "config_file" parameter to the path of the config file
// as a command line argument to the container
func WithConfigFile(cfg string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		cfgFile := testcontainers.ContainerFile{
			HostFilePath:      cfg,
			ContainerFilePath: "/etc/postgresql.conf",
			FileMode:          0o755,
		}

		req.Files = append(req.Files, cfgFile)
		req.Cmd = append(req.Cmd, "-c", "config_file=/etc/postgresql.conf")
	}
}

// WithDatabase sets the initial database to be created when the container starts
// It can be used to define a different name for the default database that is created when the image is first started.
// If it is not specified, then the value of WithUser will be used.
func WithDatabase(dbName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["POSTGRES_DB"] = dbName
	}
}

// WithInitScripts sets the init scripts to be run when the container starts
func WithInitScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		initScripts := []testcontainers.ContainerFile{}
		for _, script := range scripts {
			cf := testcontainers.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
				FileMode:          0o755,
			}
			initScripts = append(initScripts, cf)
		}
		req.Files = append(req.Files, initScripts...)
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is required for you to use the PostgreSQL image. It must not be empty or undefined.
// This environment variable sets the superuser password for PostgreSQL.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["POSTGRES_PASSWORD"] = password
	}
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a user and its password.
// It will create the specified user with superuser power and a database with the same name.
// If it is not specified, then the default user of postgres will be used.
func WithUsername(user string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		if user == "" {
			user = defaultUser
		}

		req.Env["POSTGRES_USER"] = user
	}
}

// RunContainer creates an instance of the postgres container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: defaultPostgresImage,
		Env: map[string]string{
			"POSTGRES_USER":     defaultUser,
			"POSTGRES_PASSWORD": defaultPassword,
			"POSTGRES_DB":       defaultUser, // defaults to the user name
		},
		ExposedPorts: []string{"5432/tcp"},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	user := req.Env["POSTGRES_USER"]
	password := req.Env["POSTGRES_PASSWORD"]
	dbName := req.Env["POSTGRES_DB"]

	return &PostgresContainer{Container: container, dbName: dbName, password: password, user: user}, nil
}

type snapshotConfig struct {
	snapshotName string
}

// SnapshotOption is the type for passing options to the snapshot function of the database
type SnapshotOption func(container *snapshotConfig) *snapshotConfig

// WithSnapshotName adds a specific name to the snapshot database created from the main database defined on the
// container. The snapshot must not have the same name as your main database, otherwise it will be overwritten
func WithSnapshotName(name string) SnapshotOption {
	return func(cfg *snapshotConfig) *snapshotConfig {
		cfg.snapshotName = name
		return cfg
	}
}

// Snapshot takes a snapshot of the current state of the database as a template, which can then be restored using
// the Restore method. By default, the snapshot will be created under a database called migrated_template, you can
// customize the snapshot name with the options.
// If a snapshot already exists under the given/default name, it will be overwritten with the new snapshot.
func (c *PostgresContainer) Snapshot(ctx context.Context, opts ...SnapshotOption) error {
	config := &snapshotConfig{}
	for _, opt := range opts {
		config = opt(config)
	}

	snapshotName := defaultSnapshotName
	if config.snapshotName != "" {
		snapshotName = config.snapshotName
	}

	// execute the commands to create the snapshot, in order
	cmds := []string{
		// Drop the snapshot database if it already exists
		fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, snapshotName),
		// Create a copy of the database to another database to use as a template now that it was fully migrated
		fmt.Sprintf(`CREATE DATABASE "%s" WITH TEMPLATE "%s" OWNER "%s"`, snapshotName, c.dbName, c.user),
		// Snapshot the template database so we can restore it onto our original database going forward
		fmt.Sprintf(`ALTER DATABASE "%s" WITH is_template = TRUE`, snapshotName),
	}

	for _, cmd := range cmds {
		_, _, err := c.Exec(ctx, []string{"psql", "-U", c.user, "-c", cmd})
		if err != nil {
			return err
		}
	}

	c.snapshotName = snapshotName

	return nil
}

// Restore will restore the database to a specific snapshot. By default, it will restore the last snapshot taken on the
// database by the Snapshot method. If a snapshot name is provided, it will instead try to restore the snapshot by name.
func (c *PostgresContainer) Restore(ctx context.Context, opts ...SnapshotOption) error {
	config := &snapshotConfig{}
	for _, opt := range opts {
		config = opt(config)
	}

	snapshotName := c.snapshotName
	if config.snapshotName != "" {
		snapshotName = config.snapshotName
	}

	// execute the commands to restore the snapshot, in order
	cmds := []string{
		// Drop the entire database by connecting to the postgres global database
		fmt.Sprintf(`DROP DATABASE "%s" with (FORCE)`, c.dbName),
		// Then restore the previous snapshot
		fmt.Sprintf(`CREATE DATABASE "%s" WITH TEMPLATE "%s" OWNER "%s"`, c.dbName, snapshotName, c.user),
	}

	for _, cmd := range cmds {
		_, _, err := c.Exec(ctx, []string{"psql", "-U", c.user, "-d", "postgres", "-c", cmd})
		if err != nil {
			return err
		}
	}

	return nil
}
