package postgres

import (
	"fmt"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	env map[string]string
	// SQLDriverName is the name of the SQL driver to use.
	SQLDriverName string
	Snapshot      string
}

func defaultOptions() options {
	return options{
		env: map[string]string{
			"POSTGRES_USER":     defaultUser,
			"POSTGRES_PASSWORD": defaultPassword,
			"POSTGRES_DB":       defaultUser, // defaults to the user name
		},
		SQLDriverName: "postgres",
		Snapshot:      defaultSnapshotName,
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Redpanda container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithSQLDriver sets the SQL driver to use for the container.
// It is passed to sql.Open() to connect to the database when making or restoring snapshots.
// This can be set if your app imports a different postgres driver, f.ex. "pgx"
func WithSQLDriver(driver string) Option {
	return func(o *options) error {
		o.SQLDriverName = driver
		return nil
	}
}

// WithConfigFile sets the config file to be used for the postgres container
// It will also set the "config_file" parameter to the path of the config file
// as a command line argument to the container
func WithConfigFile(cfg string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cfgFile := testcontainers.ContainerFile{
			HostFilePath:      cfg,
			ContainerFilePath: "/etc/postgresql.conf",
			FileMode:          0o755,
		}

		req.Files = append(req.Files, cfgFile)
		req.Cmd = append(req.Cmd, "-c", "config_file=/etc/postgresql.conf")

		return nil
	}
}

// WithDatabase sets the initial database to be created when the container starts
// It can be used to define a different name for the default database that is created when the image is first started.
// If it is not specified, then the value of WithUser will be used.
func WithDatabase(dbName string) Option {
	return func(o *options) error {
		o.env["POSTGRES_DB"] = dbName
		return nil
	}
}

// WithInitScripts sets the init scripts to be run when the container starts.
// These init scripts will be executed in sorted name order as defined by the container's current locale, which defaults to en_US.utf8.
// If you need to run your scripts in a specific order, consider using `WithOrderedInitScripts` instead.
func WithInitScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	containerFiles := []testcontainers.ContainerFile{}
	for _, script := range scripts {
		initScript := testcontainers.ContainerFile{
			HostFilePath:      script,
			ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
			FileMode:          0o755,
		}
		containerFiles = append(containerFiles, initScript)
	}

	return testcontainers.WithFiles(containerFiles...)
}

// WithOrderedInitScripts sets the init scripts to be run when the container starts.
// The scripts will be run in the order that they are provided in this function.
func WithOrderedInitScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	containerFiles := []testcontainers.ContainerFile{}
	for idx, script := range scripts {
		initScript := testcontainers.ContainerFile{
			HostFilePath:      script,
			ContainerFilePath: "/docker-entrypoint-initdb.d/" + fmt.Sprintf("%03d-%s", idx, filepath.Base(script)),
			FileMode:          0o755,
		}
		containerFiles = append(containerFiles, initScript)
	}

	return testcontainers.WithFiles(containerFiles...)
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is required for you to use the PostgreSQL image. It must not be empty or undefined.
// This environment variable sets the superuser password for PostgreSQL.
func WithPassword(password string) Option {
	return func(o *options) error {
		o.env["POSTGRES_PASSWORD"] = password
		return nil
	}
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a user and its password.
// It will create the specified user with superuser power and a database with the same name.
// If it is not specified, then the default user of postgres will be used.
func WithUsername(user string) Option {
	return func(o *options) error {
		if user == "" {
			user = defaultUser
		}

		o.env["POSTGRES_USER"] = user
		return nil
	}
}
