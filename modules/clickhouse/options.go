package clickhouse

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go"
)

//go:embed mounts/zk_config.xml.tpl
var zookeeperConfigTpl string

// ZookeeperOptions arguments for zookeeper in clickhouse
type ZookeeperOptions struct {
	Host, Port string
}

// renderZookeeperConfig generate default zookeeper configuration for clickhouse
func renderZookeeperConfig(settings ZookeeperOptions) ([]byte, error) {
	tpl, err := template.New("bootstrap.yaml").Parse(zookeeperConfigTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse zookeeper config file template: %w", err)
	}

	var bootstrapConfig bytes.Buffer
	if err := tpl.Execute(&bootstrapConfig, settings); err != nil {
		return nil, fmt.Errorf("failed to render zookeeper bootstrap config template: %w", err)
	}

	return bootstrapConfig.Bytes(), nil
}

// WithZookeeper pass a config to connect clickhouse with zookeeper and make clickhouse as cluster.
// It creates a temporary file in the host filesystem with the config and copies it to the container
// at /etc/clickhouse-server/config.d/zookeeper_config.xml. This file is not cleaned up automatically,
// and it's removed by the OS.
func WithZookeeper(host, port string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		f, err := os.CreateTemp("", "clickhouse-tc-config-")
		if err != nil {
			return fmt.Errorf("temporary file: %w", err)
		}

		defer f.Close()

		// write data to the temporary file
		data, err := renderZookeeperConfig(ZookeeperOptions{Host: host, Port: port})
		if err != nil {
			return fmt.Errorf("zookeeper config: %w", err)
		}
		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("write zookeeper config: %w", err)
		}
		cf := testcontainers.ContainerFile{
			HostFilePath:      f.Name(),
			ContainerFilePath: "/etc/clickhouse-server/config.d/zookeeper_config.xml",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

// WithInitScripts sets the init scripts to be run when the container starts
func WithInitScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
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

		return nil
	}
}

// WithConfigFile sets the XML config file to be used for the clickhouse container.
// The file is copied to the container at /etc/clickhouse-server/config.d/config.xml,
// which is the default location for ClickHouse config files.
func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/clickhouse-server/config.d/config.xml",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

// WithYamlConfigFile sets the YAML config file to be used for the clickhouse container
// The file is copied to the container at /etc/clickhouse-server/config.d/config.yaml,
// which is the default location for ClickHouse YAML config files.
func WithYamlConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/clickhouse-server/config.d/config.yaml",
			FileMode:          0o755,
		}
		req.Files = append(req.Files, cf)

		return nil
	}
}

// WithDatabase sets the initial database to be created when the container starts
// It can be used to define a different name for the default database that is created when the image is first started.
// If it is not specified, then the default value("clickhouse") will be used.
func WithDatabase(dbName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["CLICKHOUSE_DB"] = dbName

		return nil
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is required for you to use the ClickHouse image. It must not be empty or undefined.
// This environment variable sets the password for ClickHouse.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["CLICKHOUSE_PASSWORD"] = password

		return nil
	}
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a user and its password.
// It will create the specified user with superuser power.
// If it is not specified, then the default user of clickhouse will be used.
func WithUsername(user string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if user == "" {
			user = defaultUser
		}

		req.Env["CLICKHOUSE_USER"] = user

		return nil
	}
}
