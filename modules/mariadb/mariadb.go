package mariadb

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	rootUser            = "root"
	defaultUser         = "test"
	defaultPassword     = "test"
	defaultDatabaseName = "test"
)

// MariaDBContainer represents the MariaDB container type used in the module
type MariaDBContainer struct {
	testcontainers.Container
	username string
	password string
	database string
}

// WithDefaultCredentials applies the default credentials to the container request.
// It will look up for MARIADB environment variables.
func WithDefaultCredentials() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		username := req.Env["MARIADB_USER"]
		password := req.Env["MARIADB_PASSWORD"]
		if strings.EqualFold(rootUser, username) {
			delete(req.Env, "MARIADB_USER")
		}

		if len(password) != 0 && password != "" {
			req.Env["MARIADB_ROOT_PASSWORD"] = password
		} else if strings.EqualFold(rootUser, username) {
			req.Env["MARIADB_ALLOW_EMPTY_ROOT_PASSWORD"] = "yes"
			delete(req.Env, "MARIADB_PASSWORD")
		}

		return nil
	}
}

// https://github.com/docker-library/docs/tree/master/mariadb#environment-variables
// From tag 10.2.38, 10.3.29, 10.4.19, 10.5.10 onwards, and all 10.6 and later tags,
// the MARIADB_* equivalent variables are provided. MARIADB_* variants will always be
// used in preference to MYSQL_* variants.
func withMySQLEnvVars() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		// look up for MARIADB environment variables and apply the same to MYSQL
		for k, v := range req.Env {
			if strings.HasPrefix(k, "MARIADB_") {
				// apply the same value to the MYSQL environment variables
				mysqlEnvVar := strings.ReplaceAll(k, "MARIADB_", "MYSQL_")
				req.Env[mysqlEnvVar] = v
			}
		}

		return nil
	}
}

func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MARIADB_USER"] = username

		return nil
	}
}

func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MARIADB_PASSWORD"] = password

		return nil
	}
}

func WithDatabase(database string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MARIADB_DATABASE"] = database

		return nil
	}
}

func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/mysql/conf.d/my.cnf",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

func WithScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		var initScripts []testcontainers.ContainerFile
		for _, script := range scripts {
			cf := testcontainers.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
				FileMode:          0o755,
			}
			initScripts = append(initScripts, cf)
		}
		req.Files = append(req.Files, initScripts...)

		return nil
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the MariaDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MariaDBContainer, error) {
	return Run(ctx, "mariadb:11.0.3", opts...)
}

// Run creates an instance of the MariaDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MariaDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"3306/tcp", "33060/tcp"},
		Env: map[string]string{
			"MARIADB_USER":     defaultUser,
			"MARIADB_PASSWORD": defaultPassword,
			"MARIADB_DATABASE": defaultDatabaseName,
		},
		WaitingFor: wait.ForLog("port: 3306  mariadb.org binary distribution"),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	opts = append(opts, WithDefaultCredentials())

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	// Apply MySQL environment variables after user customization
	// In future releases of MariaDB, they could remove the MYSQL_* environment variables
	// at all. Then we can remove this customization.
	if err := withMySQLEnvVars().Customize(&genericContainerReq); err != nil {
		return nil, err
	}

	username, ok := req.Env["MARIADB_USER"]
	if !ok {
		username = rootUser
	}
	password := req.Env["MARIADB_PASSWORD"]

	if len(password) == 0 && password == "" && !strings.EqualFold(rootUser, username) {
		return nil, fmt.Errorf("empty password can be used only with the root user")
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *MariaDBContainer
	if container != nil {
		c = &MariaDBContainer{
			Container: container,
			username:  username,
			password:  password,
			database:  req.Env["MARIADB_DATABASE"],
		}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// MustConnectionString panics if the address cannot be determined.
func (c *MariaDBContainer) MustConnectionString(ctx context.Context, args ...string) string {
	addr, err := c.ConnectionString(ctx, args...)
	if err != nil {
		panic(err)
	}
	return addr
}

func (c *MariaDBContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	containerPort, err := c.MappedPort(ctx, "3306/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	extraArgs := ""
	if len(args) > 0 {
		extraArgs = strings.Join(args, "&")
	}
	if extraArgs != "" {
		extraArgs = "?" + extraArgs
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", c.username, c.password, host, containerPort.Port(), c.database, extraArgs)
	return connectionString, nil
}
