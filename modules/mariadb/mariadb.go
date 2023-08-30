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

// defaultImage {
const defaultImage = "mariadb:11.0.3"

// }

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
	return func(req *testcontainers.GenericContainerRequest) {
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
	}
}

// https://github.com/docker-library/docs/tree/master/mariadb#environment-variables
// From tag 10.2.38, 10.3.29, 10.4.19, 10.5.10 onwards, and all 10.6 and later tags,
// the MARIADB_* equivalent variables are provided. MARIADB_* variants will always be
// used in preference to MYSQL_* variants.
func withMySQLEnvVars() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		// look up for MARIADB environment variables and apply the same to MYSQL
		for k, v := range req.Env {
			if strings.HasPrefix(k, "MARIADB_") {
				// apply the same value to the MYSQL environment variables
				mysqlEnvVar := strings.ReplaceAll(k, "MARIADB_", "MYSQL_")
				req.Env[mysqlEnvVar] = v
			}
		}
	}
}

func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MARIADB_USER"] = username
	}
}

func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MARIADB_PASSWORD"] = password
	}
}

func WithDatabase(database string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MARIADB_DATABASE"] = database
	}
}

func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/mysql/conf.d/my.cnf",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)
	}
}

func WithScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
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
	}
}

// RunContainer creates an instance of the MariaDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MariaDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultImage,
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
		opt.Customize(&genericContainerReq)
	}

	// Apply MySQL environment variables after user customization
	// In future releases of MariaDB, they could remove the MYSQL_* environment variables
	// at all. Then we can remove this customization.
	withMySQLEnvVars().Customize(&genericContainerReq)

	username, ok := req.Env["MARIADB_USER"]
	if !ok {
		username = rootUser
	}
	password := req.Env["MARIADB_PASSWORD"]

	if len(password) == 0 && password == "" && !strings.EqualFold(rootUser, username) {
		return nil, fmt.Errorf("empty password can be used only with the root user")
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	database := req.Env["MARIADB_DATABASE"]

	return &MariaDBContainer{container, username, password, database}, nil
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
