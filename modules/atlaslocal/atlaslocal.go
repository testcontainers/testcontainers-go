package atlaslocal

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Container represents the MongoDBAtlasLocal container type used in the module
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the MongoDBAtlasLocal container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForAll(wait.ForHealthCheck()),
		Env:          map[string]string{},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// AuthConfig holds the authentication configuration for the MongoDB Atlas Local
// container.
type AuthConfig struct {
	Username     string
	Password     string
	UsernameFile string
	PasswordFile string
}

// WithAuth sets the authentication configuration for the MongoDB Atlas Local
// container. This can be done by providing a username and password, or by
// providing files that contain the username and password.
func WithAuth(auth AuthConfig) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		if auth.Username != "" {
			req.Env["MONGODB_INITDB_ROOT_USERNAME"] = auth.Username
		}

		if auth.Password != "" {
			req.Env["MONGODB_INITDB_ROOT_PASSWORD"] = auth.Password
		}

		if auth.UsernameFile != "" {
			req.Env["MONGODB_INITDB_ROOT_USERNAME_FILE"] = auth.UsernameFile
		}

		if auth.PasswordFile != "" {
			req.Env["MONGODB_INITDB_ROOT_PASSWORD_FILE"] = auth.PasswordFile
		}

		return nil
	}
}

// WithDisableTelemetry opts out of telemetry for the MongoDB Atlas Local
// container by setting the DO_NOT_TRACK environment variable to 1.
func WithDisableTelemetry() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		req.Env["DO_NOT_TRACK"] = "1"

		return nil
	}
}

// WithInitDatabase sets MONGODB_INITDB_DATABASE environment variable so the
// init scripts and the default connection string target the specified database
// instead of the default "test" database.
func WithInitDatabase(database string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		req.Env["MONGODB_INITDB_DATABASE"] = database

		return nil
	}
}

// WithInitScripts mounts a directory containing .sh/.js init scripts into
// /docker-entrypoint-initdb.d so they run in alphabetical order on startup.
func WithInitScripts(scriptsDir string) testcontainers.CustomizeRequestOption {
	abs, err := filepath.Abs(scriptsDir)
	if err != nil {
		return func(req *testcontainers.GenericContainerRequest) error {
			return fmt.Errorf("get absolute path of scripts directory: %w", err)
		}
	}

	return func(req *testcontainers.GenericContainerRequest) error {
		if _, err := filepath.Abs(abs); err != nil {
			return fmt.Errorf("get absolute path of scripts directory: %w", err)
		}

		prev := req.HostConfigModifier
		req.HostConfigModifier = func(hostConfig *container.HostConfig) {
			if prev != nil {
				prev(hostConfig)
			}

			bind := fmt.Sprintf("%s:/docker-entrypoint-initdb.d:ro", abs)
			hostConfig.Binds = append(hostConfig.Binds, bind)
		}

		return nil
	}
}

// WithMongotLogFile sets the path to the file where you want to store the logs
// of Atlas Search (mongot) by setting the MONGOT_LOG_FILE environment variable.
// Note that this can be set to /dev/stdout or /dev/stderr for convenience.
func WithMongotLogFile(logFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		req.Env["MONGOT_LOG_FILE"] = logFile

		return nil
	}
}

// WithRunnerLogFile sets the path to the file where you want to store the logs
// of "runner" but setting RUNNER_LOG_FILE environment variable. Note that this
// can be set to /dev/stdout or /dev/stderr for convenience.
func WithRunnerLogFile(logFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		req.Env["RUNNER_LOG_FILE"] = logFile

		return nil
	}
}
