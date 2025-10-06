package influxdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// InfluxDbContainer represents the InfluxDB container type used in the module
//
//nolint:staticcheck //FIXME
type InfluxDbContainer struct {
	testcontainers.Container
}

// Deprecated: use Run instead
// RunContainer creates an instance of the InfluxDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*InfluxDbContainer, error) {
	return Run(ctx, "influxdb:1.8", opts...)
}

// Run creates an instance of the InfluxDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*InfluxDbContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("8086/tcp", "8088/tcp"),
		testcontainers.WithEnv(map[string]string{
			"INFLUXDB_BIND_ADDRESS":          ":8088",
			"INFLUXDB_HTTP_BIND_ADDRESS":     ":8086",
			"INFLUXDB_REPORTING_DISABLED":    "true",
			"INFLUXDB_MONITOR_STORE_ENABLED": "false",
			"INFLUXDB_HTTP_HTTPS_ENABLED":    "false",
			"INFLUXDB_HTTP_AUTH_ENABLED":     "false",
		}),
		testcontainers.WithWaitStrategy(waitForHTTPHealth()),
	}
	moduleOpts = append(moduleOpts, opts...)

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *InfluxDbContainer
	if container != nil {
		c = &InfluxDbContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("run influxdb: %w", err)
	}

	return c, nil
}

//nolint:revive,staticcheck //FIXME
func (c *InfluxDbContainer) MustConnectionUrl(ctx context.Context) string {
	connectionString, err := c.ConnectionUrl(ctx)
	if err != nil {
		panic(err)
	}
	return connectionString
}

//nolint:revive,staticcheck //FIXME
func (c *InfluxDbContainer) ConnectionUrl(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "8086/tcp", "http")
}

func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"INFLUXDB_USER": username,
	})
}

func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"INFLUXDB_PASSWORD": password,
	})
}

func WithDatabase(database string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"INFLUXDB_DATABASE": database,
	})
}

func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithFiles(testcontainers.ContainerFile{
		HostFilePath:      configFile,
		ContainerFilePath: "/etc/influxdb/influxdb.conf",
		FileMode:          0o755,
	})
}

// withV2 configures the influxdb container to be compatible with InfluxDB v2
func withV2(req *testcontainers.GenericContainerRequest, org, bucket string) error {
	if org == "" {
		return errors.New("organization name is required")
	}

	if bucket == "" {
		return errors.New("bucket name is required")
	}

	return testcontainers.WithEnv(map[string]string{
		"DOCKER_INFLUXDB_INIT_ORG":    org,
		"DOCKER_INFLUXDB_INIT_BUCKET": bucket,
		"DOCKER_INFLUXDB_INIT_MODE":   "setup", // Always setup, we wont be migrating from v1 to v2
	})(req)
}

// WithV2 configures the influxdb container to be compatible with InfluxDB v2
func WithV2(org, bucket string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		err := withV2(req, org, bucket)
		if err != nil {
			return err
		}

		return nil
	}
}

const dockerSecretPath = "/run/secrets"

func secretsPath(path string) string {
	return dockerSecretPath + "/" + path
}

// WithV2Auth configures the influxdb container to be compatible with InfluxDB v2 and sets the username and password
// for the initial user.
func WithV2Auth(org, bucket, username, password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if username == "" {
			return errors.New("username is required")
		}

		if password == "" {
			return errors.New("password is required")
		}

		err := withV2(req, org, bucket)
		if err != nil {
			return err
		}

		if req.Env["DOCKER_INFLUXDB_INIT_USERNAME_FILE"] != "" ||
			req.Env["DOCKER_INFLUXDB_INIT_PASSWORD_FILE"] != "" {
			return errors.New("username and password file already set, use either WithV2Auth or WithV2SecretsAuth")
		}

		return testcontainers.WithEnv(map[string]string{
			"DOCKER_INFLUXDB_INIT_USERNAME": username,
			"DOCKER_INFLUXDB_INIT_PASSWORD": password,
		})(req)
	}
}

// WithV2SecretsAuth configures the container to be compatible with InfluxDB v2 and sets the username and password file path
func WithV2SecretsAuth(org, bucket, usernameFile, passwordFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if usernameFile == "" {
			return errors.New("username file is required")
		}

		if passwordFile == "" {
			return errors.New("password file is required")
		}

		if req.Env["DOCKER_INFLUXDB_INIT_USERNAME"] != "" ||
			req.Env["DOCKER_INFLUXDB_INIT_PASSWORD"] != "" {
			return errors.New("username and password already set, use either WithV2Auth or WithV2SecretsAuth")
		}

		err := withV2(req, org, bucket)
		if err != nil {
			return err
		}

		return testcontainers.WithEnv(map[string]string{
			"DOCKER_INFLUXDB_INIT_USERNAME_FILE": secretsPath(usernameFile),
			"DOCKER_INFLUXDB_INIT_PASSWORD_FILE": secretsPath(passwordFile),
		})(req)
	}
}

// WithV2Retention configures the default bucket's retention
func WithV2Retention(retention time.Duration) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if retention == 0 {
			return errors.New("retention is required")
		}

		return testcontainers.WithEnv(map[string]string{
			"DOCKER_INFLUXDB_INIT_RETENTION": retention.String(),
		})(req)
	}
}

// WithV2AdminToken sets the admin token for the influxdb container
func WithV2AdminToken(token string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if token == "" {
			return errors.New("admin token is required")
		}

		if req.Env["DOCKER_INFLUXDB_INIT_ADMIN_TOKEN_FILE"] != "" {
			return errors.New("admin token file already set, use either WithV2AdminToken or WithV2SecretsAdminToken")
		}

		return testcontainers.WithEnv(map[string]string{
			"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN": token,
		})(req)
	}
}

// WithV2SecretsAdminToken sets the admin token for the influxdb container using a file
func WithV2SecretsAdminToken(tokenFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if tokenFile == "" {
			return errors.New("admin token file is required")
		}

		if req.Env["DOCKER_INFLUXDB_INIT_ADMIN_TOKEN"] != "" {
			return errors.New("admin token already set, use either WithV2AdminToken or WithV2SecretsAdminToken")
		}

		return testcontainers.WithEnv(map[string]string{
			"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN_FILE": secretsPath(tokenFile),
		})(req)
	}
}

// WithInitDb returns a request customizer that initialises the database using the file `docker-entrypoint-initdb.d`
// located in `srcPath` directory.
//
//nolint:staticcheck //FIXME
func WithInitDb(srcPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if err := testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      path.Join(srcPath, "docker-entrypoint-initdb.d"),
			ContainerFilePath: "/",
			FileMode:          0o755,
		})(req); err != nil {
			return err
		}

		return testcontainers.WithAdditionalWaitStrategy(
			wait.ForLog("Server shutdown completed"),
			waitForHTTPHealth(),
		)(req)
	}
}

func waitForHTTPHealth() *wait.HTTPStrategy {
	return wait.ForHTTP("/health").
		WithResponseMatcher(func(body io.Reader) bool {
			decoder := json.NewDecoder(body)
			r := struct {
				Status string `json:"status"`
			}{}
			if err := decoder.Decode(&r); err != nil {
				return false
			}
			return r.Status == "pass"
		})
}
