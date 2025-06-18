package mariadb

import (
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	env map[string]string
}

func defaultOptions() options {
	return options{
		env: map[string]string{
			"MARIADB_USER":     defaultUser,
			"MARIADB_PASSWORD": defaultPassword,
			"MARIADB_DATABASE": defaultDatabaseName,
		},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the MariaDB container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithDefaultCredentials applies the default credentials to the container request.
// It will look up for MARIADB environment variables.
func WithDefaultCredentials() Option {
	return func(o *options) error {
		username := o.env["MARIADB_USER"]
		password := o.env["MARIADB_PASSWORD"]
		if strings.EqualFold(rootUser, username) {
			delete(o.env, "MARIADB_USER")
		}

		if len(password) != 0 && password != "" {
			o.env["MARIADB_ROOT_PASSWORD"] = password
		} else if strings.EqualFold(rootUser, username) {
			o.env["MARIADB_ALLOW_EMPTY_ROOT_PASSWORD"] = "yes"
			delete(o.env, "MARIADB_PASSWORD")
		}

		return nil
	}
}

// https://github.com/docker-library/docs/tree/master/mariadb#environment-variables
// From tag 10.2.38, 10.3.29, 10.4.19, 10.5.10 onwards, and all 10.6 and later tags,
// the MARIADB_* equivalent variables are provided. MARIADB_* variants will always be
// used in preference to MYSQL_* variants.
func withMySQLEnvVars() Option {
	return func(o *options) error {
		// look up for MARIADB environment variables and apply the same to MYSQL
		for k, v := range o.env {
			if strings.HasPrefix(k, "MARIADB_") {
				// apply the same value to the MYSQL environment variables
				mysqlEnvVar := strings.ReplaceAll(k, "MARIADB_", "MYSQL_")
				o.env[mysqlEnvVar] = v
			}
		}

		return nil
	}
}

func WithUsername(username string) Option {
	return func(o *options) error {
		o.env["MARIADB_USER"] = username

		return nil
	}
}

func WithPassword(password string) Option {
	return func(o *options) error {
		o.env["MARIADB_PASSWORD"] = password

		return nil
	}
}

func WithDatabase(database string) Option {
	return func(o *options) error {
		o.env["MARIADB_DATABASE"] = database

		return nil
	}
}

func WithConfigFile(configFile string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithFiles(testcontainers.ContainerFile{
		HostFilePath:      configFile,
		ContainerFilePath: "/etc/mysql/conf.d/my.cnf",
		FileMode:          0o755,
	})
}

func WithScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	var initScripts []testcontainers.ContainerFile
	for _, script := range scripts {
		cf := testcontainers.ContainerFile{
			HostFilePath:      script,
			ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
			FileMode:          0o755,
		}
		initScripts = append(initScripts, cf)
	}

	return testcontainers.WithFiles(initScripts...)
}
