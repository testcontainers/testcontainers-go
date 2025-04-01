package influxdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strconv"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// InfluxDbContainer represents the InfluxDB container type used in the module
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
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"8086/tcp", "8088/tcp"},
		Env: map[string]string{
			"INFLUXDB_BIND_ADDRESS":          ":8088",
			"INFLUXDB_HTTP_BIND_ADDRESS":     ":8086",
			"INFLUXDB_REPORTING_DISABLED":    "true",
			"INFLUXDB_MONITOR_STORE_ENABLED": "false",
			"INFLUXDB_HTTP_HTTPS_ENABLED":    "false",
			"INFLUXDB_HTTP_AUTH_ENABLED":     "false",
		},
		WaitingFor: waitForHTTPHealth(),
	}
	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *InfluxDbContainer
	if container != nil {
		c = &InfluxDbContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

//nolint:revive //FIXME
func (c *InfluxDbContainer) MustConnectionUrl(ctx context.Context) string {
	connectionString, err := c.ConnectionUrl(ctx)
	if err != nil {
		panic(err)
	}
	return connectionString
}

//nolint:revive //FIXME
func (c *InfluxDbContainer) ConnectionUrl(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "8086/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, containerPort.Port()), nil
}

func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["INFLUXDB_USER"] = username
		return nil
	}
}

func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["INFLUXDB_PASSWORD"] = password
		return nil
	}
}

func WithDatabase(database string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["INFLUXDB_DATABASE"] = database
		return nil
	}
}

func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/influxdb/influxdb.conf",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)
		return nil
	}
}

// InfluxDBV2Config contains the configuration for InfluxDB v2.
type InfluxDBV2Config struct {
	Username     *string // Username for the initial user
	Password     *string // Password for the initial user
	UsernameFile *string // File containing the username (e.g., /run/secrets/username)
	PasswordFile *string // File containing the password (e.g., /run/secrets/password)
	TokenFile    *string // File containing the token (e.g., /run/secrets/token)
	Org          string  // Organization name (required)
	Bucket       string  // Bucket name (required)
	Retention    *string // Retention policy (e.g., "30d" for 30 days)
	Token        *string // Admin token
	AuthEnabled  *bool   // Enable authentication (true/false); defaults to false.
}

// WithV2Env sets up the container with the environment variables compatible with InfluxDB v2.
// Allows for setting up the initial user, password, organization, bucket, retention and authentication options.
func WithV2Env(config InfluxDBV2Config) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DOCKER_INFLUXDB_INIT_ORG"] = config.Org
		req.Env["DOCKER_INFLUXDB_INIT_BUCKET"] = config.Bucket
		req.Env["DOCKER_INFLUXDB_INIT_AUTH_ENABLED"] = "false" // Disable auth by default
		req.Env["DOCKER_INFLUXDB_INIT_MODE"] = "setup"         // Always setup, we wont be migrating from v1 to v2

		if config.Username != nil {
			req.Env["DOCKER_INFLUXDB_INIT_USERNAME"] = *config.Username
		}

		if config.Password != nil {
			req.Env["DOCKER_INFLUXDB_INIT_PASSWORD"] = *config.Password
		}

		if config.Token != nil {
			req.Env["DOCKER_INFLUXDB_INIT_ADMIN_TOKEN"] = *config.Token
		}

		if config.AuthEnabled != nil {
			req.Env["DOCKER_INFLUXDB_INIT_AUTH_ENABLED"] = strconv.FormatBool(*config.AuthEnabled)
		}

		if config.Retention != nil {
			req.Env["DOCKER_INFLUXDB_INIT_RETENTION"] = *config.Retention
		}

		if config.UsernameFile != nil {
			req.Env["DOCKER_INFLUXDB_INIT_USERNAME_FILE"] = *config.UsernameFile
		}

		if config.PasswordFile != nil {
			req.Env["DOCKER_INFLUXDB_INIT_PASSWORD_FILE"] = *config.PasswordFile
		}

		if config.TokenFile != nil {
			req.Env["DOCKER_INFLUXDB_INIT_ADMIN_TOKEN_FILE"] = *config.TokenFile
		}

		return nil
	}
}

// WithInitDb returns a request customizer that initialises the database using the file `docker-entrypoint-initdb.d`
// located in `srcPath` directory.
func WithInitDb(srcPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      path.Join(srcPath, "docker-entrypoint-initdb.d"),
			ContainerFilePath: "/",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		req.WaitingFor = wait.ForAll(
			wait.ForLog("Server shutdown completed"),
			waitForHTTPHealth(),
		)
		return nil
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
