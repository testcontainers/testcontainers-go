package cockroachdb

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

// customizer is an interface for customizing a CockroachDB container.
type customizer interface {
	customize(*CockroachDBContainer) error
}

// WithDatabase sets the name of the database to use.
// This will be ignored if data exists in the `/cockroach/cockroach-data` directory within the container.
func WithDatabase(database string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[envDatabase] = database
		return nil
	}
}

// WithUser creates & sets the user to connect as.
// This will be ignored if data exists in the `/cockroach/cockroach-data` directory within the container.
func WithUser(user string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if user != defaultUser && req.Env[envOptionTLS] == "true" {
			return fmt.Errorf("unsupported user %q with TLS, use %q", user, defaultUser)
		}

		req.Env[envUser] = user
		return nil
	}
}

// WithPassword sets the password when using password authentication.
// This will be ignored if data exists in the `/cockroach/cockroach-data` directory within the container.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[envPassword] = password
		if password != "" {
			req.Cmd = append(req.Cmd, "--accept-sql-without-tls")
		}

		return nil
	}
}

// WithStoreSize sets the amount of available [in-memory storage].
//
// [in-memory storage]: https://www.cockroachlabs.com/docs/stable/cockroach-start#store
func WithStoreSize(size string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		for i, cmd := range req.Cmd {
			if strings.HasPrefix(cmd, memStorageFlag) {
				req.Cmd[i] = memStorageFlag + size
				return nil
			}
		}

		// Wasn't found, add it.
		req.Cmd = append(req.Cmd, memStorageFlag+size)

		return nil
	}
}

// WithNoClusterDefaults disables the default cluster settings script.
//
// Without this option Cockroach containers run `data/cluster-defaults.sql` on startup
// which configures the settings recommended by Cockroach Labs for [local testing clusters]
// unless data exists in the `/cockroach/cockroach-data` directory within the container.
//
// [local testing clusters]: https://www.cockroachlabs.com/docs/stable/local-testing
func WithNoClusterDefaults() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		for i, file := range req.Files {
			if _, ok := file.Reader.(*defaultsReader); ok && file.ContainerFilePath == clusterDefaultsContainerFile {
				req.Files = append(req.Files[:i], req.Files[i+1:]...)
				return nil
			}
		}

		return nil
	}
}

// WithInitScripts adds the given scripts to those automatically run when the container starts.
// These will be ignored if data exists in the `/cockroach/cockroach-data` directory within the container.
func WithInitScripts(scripts ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		files := make([]testcontainers.ContainerFile, len(scripts))
		for i, script := range scripts {
			files[i] = testcontainers.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: initDBPath + filepath.Base(script),
				FileMode:          0o644,
			}
		}
		req.Files = append(req.Files, files...)

		return nil
	}
}
