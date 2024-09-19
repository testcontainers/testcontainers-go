package influxdb

import (
	"context"
	"fmt"
	"path"
	"strings"

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
		WaitingFor: wait.ForListeningPort("8086/tcp"),
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

	hasInitDb := false

	for _, f := range genericContainerReq.Files {
		if f.ContainerFilePath == "/" && strings.HasSuffix(f.HostFilePath, "docker-entrypoint-initdb.d") {
			// Init service in container will start influxdb, run scripts in docker-entrypoint-initdb.d and then
			// terminate the influxdb server, followed by restart of influxdb.  This is tricky to wait for, and
			// in this case, we are assuming that data was added by init script, so we then look for an
			// "Open shard" which is the last thing that happens before the server is ready to accept connections.
			// This is probably different for InfluxDB 2.x, but that is left as an exercise for the reader.
			strategies := []wait.Strategy{
				genericContainerReq.WaitingFor,
				wait.ForLog("influxdb init process in progress..."),
				wait.ForLog("Server shutdown completed"),
				wait.ForLog("Opened shard"),
			}
			genericContainerReq.WaitingFor = wait.ForAll(strategies...)
			hasInitDb = true
			break
		}
	}

	if !hasInitDb {
		if lastIndex := strings.LastIndex(genericContainerReq.Image, ":"); lastIndex != -1 {
			tag := genericContainerReq.Image[lastIndex+1:]
			if tag == "latest" || tag[0] == '2' {
				genericContainerReq.WaitingFor = wait.ForLog(`Listening log_id=[0-9a-zA-Z_~]+ service=tcp-listener transport=http`).AsRegexp()
			}
		} else {
			genericContainerReq.WaitingFor = wait.ForLog("Listening for signals")
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

// WithInitDb will copy a 'docker-entrypoint-initdb.d' directory to the container.
// The secPath is the path to the directory on the host machine.
// The directory will be copied to the root of the container.
func WithInitDb(srcPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      path.Join(srcPath, "docker-entrypoint-initdb.d"),
			ContainerFilePath: "/",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)
		return nil
	}
}
