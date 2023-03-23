package postgres

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const defaultUser = "postgres"
const defaultPassword = "postgres"
const defaultPostgresImage = "docker.io/postgres:11-alpine"

// PostgresContainer represents the postgres container type used in the module
type PostgresContainer struct {
	testcontainers.Container
	dbName   string
	user     string
	password string
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

	extraArgs := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "sslmode=") {
			continue
		}
		extraArgs += " " + arg
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable %s", host, containerPort.Port(), c.user, c.password, c.dbName, extraArgs)
	return connStr, nil
}

// PostgresContainerOption is a function that configures the postgres container, affecting the container request
type PostgresContainerOption func(req *testcontainers.ContainerRequest)

// WithWaitStrategy sets the wait strategy for the postgres container
func WithWaitStrategy(strategies ...wait.Strategy) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(1 * time.Minute)
	}
}

// WithImage sets the image to be used for the postgres container
func WithImage(image string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if image == "" {
			image = defaultPostgresImage
		}

		req.Image = image
	}
}

// WithConfigFile sets the config file to be used for the postgres container
// It will also set the "config_file" parameter to the path of the config file
// as a command line argument to the container
func WithConfigFile(cfg string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		cfgFile := testcontainers.ContainerFile{
			HostFilePath:      cfg,
			ContainerFilePath: "/etc/postgresql.conf",
			FileMode:          0755,
		}

		req.Files = append(req.Files, cfgFile)
		req.Cmd = append(req.Cmd, "-c", "config_file=/etc/postgresql.conf")
	}

}

// WithDatabase sets the initial database to be created when the container starts
// It can be used to define a different name for the default database that is created when the image is first started.
// If it is not specified, then the value of WithUser will be used.
func WithDatabase(dbName string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["POSTGRES_DB"] = dbName
	}
}

// WithInitScripts sets the init scripts to be run when the container starts
func WithInitScripts(scripts ...string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		initScripts := []testcontainers.ContainerFile{}
		for _, script := range scripts {
			cf := testcontainers.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
				FileMode:          0755,
			}
			initScripts = append(initScripts, cf)
		}
		req.Files = append(req.Files, initScripts...)
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is required for you to use the PostgreSQL image. It must not be empty or undefined.
// This environment variable sets the superuser password for PostgreSQL.
func WithPassword(password string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["POSTGRES_PASSWORD"] = password
	}
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a user and its password.
// It will create the specified user with superuser power and a database with the same name.
// If it is not specified, then the default user of postgres will be used.
func WithUsername(user string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		if user == "" {
			user = defaultUser
		}

		req.Env["POSTGRES_USER"] = user
	}
}

// StartContainer creates an instance of the postgres container type
func StartContainer(ctx context.Context, opts ...PostgresContainerOption) (*PostgresContainer, error) {
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

	for _, opt := range opts {
		opt(&req)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	user := req.Env["POSTGRES_USER"]
	password := req.Env["POSTGRES_PASSWORD"]
	dbName := req.Env["POSTGRES_DB"]

	return &PostgresContainer{Container: container, dbName: dbName, password: password, user: user}, nil
}
