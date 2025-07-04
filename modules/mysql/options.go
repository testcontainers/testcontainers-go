package mysql

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
			"MYSQL_USER":     defaultUser,
			"MYSQL_PASSWORD": defaultPassword,
			"MYSQL_DATABASE": defaultDatabaseName,
		},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the MSSQL container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

func WithDefaultCredentials() Option {
	return func(o *options) error {
		username := o.env["MYSQL_USER"]
		password := o.env["MYSQL_PASSWORD"]
		if strings.EqualFold(rootUser, username) {
			delete(o.env, "MYSQL_USER")
		}
		if len(password) != 0 && password != "" {
			o.env["MYSQL_ROOT_PASSWORD"] = password
		} else if strings.EqualFold(rootUser, username) {
			o.env["MYSQL_ALLOW_EMPTY_PASSWORD"] = "yes"
			delete(o.env, "MYSQL_PASSWORD")
		}

		return nil
	}
}

func WithUsername(username string) Option {
	return func(o *options) error {
		o.env["MYSQL_USER"] = username

		return nil
	}
}

func WithPassword(password string) Option {
	return func(o *options) error {
		o.env["MYSQL_PASSWORD"] = password

		return nil
	}
}

func WithDatabase(database string) Option {
	return func(o *options) error {
		o.env["MYSQL_DATABASE"] = database

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
