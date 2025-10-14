package mariadb

import (
	"context"
	"errors"
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
	return testcontainers.WithEnv(map[string]string{
		"MARIADB_USER": username,
	})
}

func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"MARIADB_PASSWORD": password,
	})
}

func WithDatabase(database string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"MARIADB_DATABASE": database,
	})
}

func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithFiles(testcontainers.ContainerFile{
		HostFilePath:      configFile,
		ContainerFilePath: "/etc/mysql/conf.d/my.cnf",
		FileMode:          0o755,
	})
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

		if err := testcontainers.WithFiles(initScripts...)(req); err != nil {
			return err
		}

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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("3306/tcp", "33060/tcp"),
		testcontainers.WithEnv(map[string]string{
			"MARIADB_USER":     defaultUser,
			"MARIADB_PASSWORD": defaultPassword,
			"MARIADB_DATABASE": defaultDatabaseName,
		}),
		testcontainers.WithWaitStrategy(wait.ForLog("port: 3306  mariadb.org binary distribution")),
	}

	moduleOpts = append(moduleOpts, opts...)
	moduleOpts = append(moduleOpts, WithDefaultCredentials())

	// Apply MySQL environment variables after user customization
	// In future releases of MariaDB, they could remove the MYSQL_* environment variables
	// at all. Then we can remove this customization.
	moduleOpts = append(moduleOpts, withMySQLEnvVars())

	// validate credentials before running the container
	moduleOpts = append(moduleOpts, validateCredentials())

	var c *MariaDBContainer
	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	if ctr != nil {
		c = &MariaDBContainer{Container: ctr, username: rootUser}
	}
	if err != nil {
		return c, fmt.Errorf("run mariadb: %w", err)
	}

	// Inspect the container to get environment variables
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect mariadb: %w", err)
	}

	foundUser, foundPass, foundDB := false, false, false
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "MARIADB_USER="); ok {
			c.username, foundUser = v, true
		}
		if v, ok := strings.CutPrefix(env, "MARIADB_PASSWORD="); ok {
			c.password, foundPass = v, true
		}
		if v, ok := strings.CutPrefix(env, "MARIADB_DATABASE="); ok {
			c.database, foundDB = v, true
		}

		if foundUser && foundPass && foundDB {
			break
		}
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
	endpoint, err := c.PortEndpoint(ctx, "3306/tcp", "")
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

	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s%s", c.username, c.password, endpoint, c.database, extraArgs)
	return connectionString, nil
}

// validateCredentials validates the credentials to ensure that an empty password
// is only used with the root user.
func validateCredentials() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		username, ok := req.Env["MARIADB_USER"]
		if !ok {
			username = rootUser
		}
		password := req.Env["MARIADB_PASSWORD"]

		if len(password) == 0 && password == "" && !strings.EqualFold(rootUser, username) {
			return errors.New("empty password can be used only with the root user")
		}

		return nil
	}
}
