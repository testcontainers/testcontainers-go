package cockroachdb

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// errInsecureWithPassword is returned when trying to use insecure mode with a password.
var errInsecureWithPassword = errors.New("insecure mode cannot be used with a password")

// WithDatabase sets the name of the database to create and use.
// This will be converted to lowercase as CockroachDB forces the database to be lowercase.
// The database creation will be skipped if data exists in the `/cockroach/cockroach-data` directory within the container.
func WithDatabase(database string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[envDatabase] = strings.ToLower(database)
		return nil
	}
}

// WithUser sets the name of the user to create and connect as.
// This will be converted to lowercase as CockroachDB forces the user to be lowercase.
// The user creation will be skipped if data exists in the `/cockroach/cockroach-data` directory within the container.
func WithUser(user string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[envUser] = strings.ToLower(user)
		return nil
	}
}

// WithPassword sets the password of the user to create and connect as.
// The user creation will be skipped if data exists in the `/cockroach/cockroach-data` directory within the container.
// This will error if insecure mode is enabled.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		for _, arg := range req.Cmd {
			if arg == insecureFlag {
				return errInsecureWithPassword
			}
		}

		req.Env[envPassword] = password

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
				ContainerFilePath: initDBPath + "/" + filepath.Base(script),
				FileMode:          0o644,
			}
		}
		req.Files = append(req.Files, files...)

		return nil
	}
}

// WithInsecure enables insecure mode which disables TLS.
func WithInsecure() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env[envPassword] != "" {
			return errInsecureWithPassword
		}

		req.Cmd = append(req.Cmd, insecureFlag)

		return nil
	}
}

// configure sets the CockroachDBContainer options from the given request and updates the request
// wait strategies to match the options.
// This option must be called after all the options have been applied, in order to extract
// the credentials from the environment variables and the TLS strategy from the wait strategy.
func (c *CockroachDBContainer) configure() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		// refresh the credentials from the environment variables
		c.user = req.Env[envUser]
		c.password = req.Env[envPassword]
		c.database = req.Env[envDatabase]

		var insecure bool
		if slices.Contains(req.Cmd, insecureFlag) {
			insecure = true
		}

		// Walk the wait strategies to find the TLS strategy and either remove it or
		// update the client certificate files to match the user and configure the
		// container to use the TLS strategy.
		if err := wait.Walk(&req.WaitingFor, func(strategy wait.Strategy) error {
			if cert, ok := strategy.(*wait.TLSStrategy); ok {
				if insecure {
					// If insecure mode is enabled, the certificate strategy is removed.
					return errors.Join(wait.ErrVisitRemove, wait.ErrVisitStop)
				}

				// Update the client certificate files to match the user which may have changed.
				cert.WithCert(certsDir+"/client."+c.user+".crt", certsDir+"/client."+c.user+".key")

				c.tlsStrategy = cert

				// Stop the walk as the certificate strategy has been found.
				return wait.ErrVisitStop
			}

			return nil
		}); err != nil {
			return fmt.Errorf("walk strategies: %w", err)
		}

		return nil
	}
}
