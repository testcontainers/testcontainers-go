package influxdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// InfluxDbContainer represents the MySQL container type used in the module
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
		WaitingFor: waitForHttpHealth(),
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

func (c *InfluxDbContainer) MustConnectionUrl(ctx context.Context) string {
	connectionString, err := c.ConnectionUrl(ctx)
	if err != nil {
		panic(err)
	}
	return connectionString
}

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

// WithInitDb copies a 'docker-entrypoint-initdb.d' directory from the specified host path to the root of the container.
// The `srcPath` parameter should point to the directory containing initialization files on the host.
//
// Initialization Process in the Container:
// 1. The copied 'docker-entrypoint-initdb.d' directory contains scripts that initialize the database.
// 2. On container start, InfluxDB runs, executes the scripts in 'docker-entrypoint-initdb.d', and then shuts down.
// 3. The InfluxDB server restarts automatically after initialization to make the new data available.
//
// Note: This approach assumes the initialization completes on startup and the data is properly added.
// This behavior may differ in InfluxDB 2.x and may require additional handling.
//
// Parameters:
// - srcPath: The host path to the directory containing initialization scripts.
//
// Returns:
// - testcontainers.CustomizeRequestOption: An option to customize the container request.
func WithInitDb(srcPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      path.Join(srcPath, "docker-entrypoint-initdb.d"),
			ContainerFilePath: "/",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		strategies := []wait.Strategy{
			wait.ForLog("Server shutdown completed"),
			waitForHttpHealth(),
		}
		req.WaitingFor = wait.ForAll(strategies...)
		return nil
	}
}

func waitForHttpHealth() *wait.HTTPStrategy {
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
