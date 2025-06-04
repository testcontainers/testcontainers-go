package dolt

import (
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	env   map[string]string
	files []testcontainers.ContainerFile
}

func defaultOptions() options {
	return options{
		env: map[string]string{
			"DOLT_USER":     defaultUser,
			"DOLT_PASSWORD": defaultPassword,
			"DOLT_DATABASE": defaultDatabaseName,
		},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Dolt container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

func WithDefaultCredentials() Option {
	return func(o *options) error {
		username := o.env["DOLT_USER"]
		if strings.EqualFold(rootUser, username) {
			delete(o.env, "DOLT_USER")
			delete(o.env, "DOLT_PASSWORD")
		}

		return nil
	}
}

func WithUsername(username string) Option {
	return func(o *options) error {
		o.env["DOLT_USER"] = username
		return nil
	}
}

func WithPassword(password string) Option {
	return func(o *options) error {
		o.env["DOLT_PASSWORD"] = password
		return nil
	}
}

func WithDoltCredsPublicKey(key string) Option {
	return func(o *options) error {
		o.env["DOLT_CREDS_PUB_KEY"] = key
		return nil
	}
}

//nolint:revive,staticcheck //FIXME
func WithDoltCloneRemoteUrl(url string) Option {
	return func(o *options) error {
		o.env["DOLT_REMOTE_CLONE_URL"] = url
		return nil
	}
}

func WithDatabase(database string) Option {
	return func(o *options) error {
		o.env["DOLT_DATABASE"] = database
		return nil
	}
}

func WithConfigFile(configFile string) Option {
	return func(o *options) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      configFile,
			ContainerFilePath: "/etc/dolt/servercfg.d/server.cnf",
			FileMode:          0o755,
		}
		o.files = append(o.files, cf)
		return nil
	}
}

func WithCredsFile(credsFile string) Option {
	return func(o *options) error {
		cf := testcontainers.ContainerFile{
			HostFilePath:      credsFile,
			ContainerFilePath: "/root/.dolt/creds/" + filepath.Base(credsFile),
			FileMode:          0o755,
		}
		o.files = append(o.files, cf)
		return nil
	}
}

func WithScripts(scripts ...string) Option {
	return func(o *options) error {
		var initScripts []testcontainers.ContainerFile
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
