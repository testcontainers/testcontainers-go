package clickhouse

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	env   map[string]string
	files []testcontainers.ContainerFile
}

func defaultOptions() options {
	return options{
		env: map[string]string{
			"CLICKHOUSE_USER":     defaultUser,
			"CLICKHOUSE_PASSWORD": defaultUser,
			"CLICKHOUSE_DB":       defaultDatabaseName,
		},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the ClickHouse container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithZookeeper pass a config to connect clickhouse with zookeeper and make clickhouse as cluster
func WithZookeeper(host, port string) Option {
	return func(o *options) error {
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
		o.files = append(o.files, cf)

		return nil
	}
}

// WithInitScripts sets the init scripts to be run when the container starts
func WithInitScripts(scripts ...string) Option {
	return func(o *options) error {
		initScripts := []testcontainers.ContainerFile{}
		for _, script := range scripts {
			cf := testcontainers.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
				FileMode:          0o755,
			}
			initScripts = append(initScripts, cf)
		}
		o.files = append(o.files, initScripts...)

		return nil
	}
}

// WithConfigFile sets the XML config file to be used for the clickhouse container
// It will also set the "configFile" parameter to the path of the config file
// as a command line argument to the container.
func WithConfigFile(configFile string) Option {
	return func(o *options) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/clickhouse-server/config.d/config.xml",
			FileMode:          0o755,
		}
		o.files = append(o.files, cf)

		return nil
	}
}

// WithConfigFile sets the YAML config file to be used for the clickhouse container
// It will also set the "configFile" parameter to the path of the config file
// as a command line argument to the container.
func WithYamlConfigFile(configFile string) Option {
	return func(o *options) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/clickhouse-server/config.d/config.yaml",
			FileMode:          0o755,
		}
		o.files = append(o.files, cf)

		return nil
	}
}

// WithDatabase sets the initial database to be created when the container starts
// It can be used to define a different name for the default database that is created when the image is first started.
// If it is not specified, then the default value("clickhouse") will be used.
func WithDatabase(dbName string) Option {
	return func(o *options) error {
		o.env["CLICKHOUSE_DB"] = dbName

		return nil
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is required for you to use the ClickHouse image. It must not be empty or undefined.
// This environment variable sets the password for ClickHouse.
func WithPassword(password string) Option {
	return func(o *options) error {
		o.env["CLICKHOUSE_PASSWORD"] = password

		return nil
	}
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a user and its password.
// It will create the specified user with superuser power.
// If it is not specified, then the default user of clickhouse will be used.
func WithUsername(user string) Option {
	return func(o *options) error {
		if user == "" {
			user = defaultUser
		}

		o.env["CLICKHOUSE_USER"] = user

		return nil
	}
}
