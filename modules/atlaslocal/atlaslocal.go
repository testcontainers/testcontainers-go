package atlaslocal

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Container represents the MongoDBAtlasLocal container type used in the module
type Container struct {
	testcontainers.Container
	username string
	password string
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
			return nil, fmt.Errorf("failed to create customized request: %w", err)
		}
	}

	if err := validateRequest(&genericContainerReq); err != nil {
		return nil, fmt.Errorf("incompatible configuration: %w", err)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	// Set the username from the username file so that it can be used in the
	// connection string method.
	c.username = genericContainerReq.Env["MONGODB_INITDB_ROOT_USERNAME"]
	if usernameFile := genericContainerReq.Env["MONGODB_INITDB_ROOT_USERNAME_FILE"]; usernameFile != "" {
		fileContent, err := os.ReadFile(usernameFile)
		if err != nil {
			return nil, fmt.Errorf("read username file: %w", err)
		}

		c.username = strings.TrimSpace(string(fileContent))
	}

	// Set the password from the password file so that it can be used in the
	// connection string method.
	c.password = genericContainerReq.Env["MONGODB_INITDB_ROOT_PASSWORD"]
	if passwordFile := genericContainerReq.Env["MONGODB_INITDB_ROOT_PASSWORD_FILE"]; passwordFile != "" {
		fileContent, err := os.ReadFile(passwordFile)
		if err != nil {
			return nil, fmt.Errorf("read password file: %w", err)
		}

		c.password = strings.TrimSpace(string(fileContent))
	}

	return c, nil
}

func validateRequest(req *testcontainers.GenericContainerRequest) error {
	username := req.Env["MONGODB_INITDB_ROOT_USERNAME"]
	password := req.Env["MONGODB_INITDB_ROOT_PASSWORD"]

	// If username or password is specified, both must be provided.
	if username != "" && password == "" || username == "" && password != "" {
		return errors.New("if you specify username or password, you must provide both of them")
	}

	usernameFile := req.Env["MONGODB_INITDB_ROOT_USERNAME_FILE"]
	passwordFile := req.Env["MONGODB_INITDB_ROOT_PASSWORD_FILE"]

	// If username file or password file is specified, both must be provided.
	if usernameFile != "" && passwordFile == "" || usernameFile == "" && passwordFile != "" {
		return errors.New("if you specify username file or password file, you must provide both of them")
	}

	// Setting credentials both inline and using files will result in an panic
	// from the container, so we short circuit here.
	if (username != "" || password != "") && (usernameFile != "" || passwordFile != "") {
		return errors.New("you cannot specify both inline credentials and files for credentials")
	}

	return nil
}

// ConnectionString returns the connection string for the MongoDB Atlas Local
// container. If you provide a username and a password, the connection string
// will also include them.
func (ctr *Container) ConnectionString(ctx context.Context) (string, error) {
	host, err := ctr.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := ctr.MappedPort(ctx, "27017")
	if err != nil {
		return "", err
	}

	uri := &url.URL{
		Scheme:   "mongodb",
		Host:     net.JoinHostPort(host, mappedPort.Port()),
		Path:     "/",
		RawQuery: "directConnection=true",
	}

	if ctr.username != "" && ctr.password != "" {
		uri.User = url.UserPassword(ctr.username, ctr.password)
	}

	return uri.String(), nil
}

// WithUsername sets the MongoDB root username by setting the
// MONGODB_INITDB_ROOT_USERNAME environment variable.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if username != "" {
			req.Env["MONGODB_INITDB_ROOT_USERNAME"] = username
		}

		return nil
	}
}

// WithPassword sets the MongoDB root password by setting the the
// MONGODB_INITDB_ROOT_PASSWORD environment variable.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if password != "" {
			req.Env["MONGODB_INITDB_ROOT_PASSWORD"] = password
		}

		return nil
	}
}

// WithUsernameFile sets the file path to source the MongoDB root username by
// setting the MONGODB_INITDB_ROOT_USERNAME_FILE environment variable. This
// function mounts the local file into the container at the same path.
func WithUsernameFile(usernameFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if usernameFile != "" {
			req.Env["MONGODB_INITDB_ROOT_USERNAME_FILE"] = usernameFile
		}

		prev := req.HostConfigModifier
		req.HostConfigModifier = func(hostConfig *container.HostConfig) {
			if prev != nil {
				prev(hostConfig)
			}

			// Mount username file.
			if usernameFile != "" {
				hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s:ro", usernameFile, usernameFile))
			}
		}

		return nil
	}
}

// WithPasswordFile sets the file path to source the MongoDB root password by
// setting the MONGODB_INITDB_ROOT_PASSWORD_FILE environment variable. This
// function mounts the local file into the container at the same path.
func WithPasswordFile(passwordFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if passwordFile != "" {
			req.Env["MONGODB_INITDB_ROOT_PASSWORD_FILE"] = passwordFile
		}

		prev := req.HostConfigModifier
		req.HostConfigModifier = func(hostConfig *container.HostConfig) {
			if prev != nil {
				prev(hostConfig)
			}

			// Mount username file.
			if passwordFile != "" {
				hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s:ro", passwordFile, passwordFile))
			}
		}

		return nil
	}
}

// WithDisableTelemetry opts out of telemetry for the MongoDB Atlas Local
// container by setting the DO_NOT_TRACK environment variable to 1.
func WithDisableTelemetry() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DO_NOT_TRACK"] = "1"

		return nil
	}
}

// WithInitDatabase sets MONGODB_INITDB_DATABASE environment variable so the
// init scripts and the default connection string target the specified database
// instead of the default "test" database.
func WithInitDatabase(database string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MONGODB_INITDB_DATABASE"] = database

		return nil
	}
}

// WithInitScripts mounts a directory containing .sh/.js init scripts into
// /docker-entrypoint-initdb.d so they run in alphabetical order on startup. If
// called multiple times, this funcion removes any prior init-scripts bind and
// uses only the latest on specified.
func WithInitScripts(scriptsDir string) testcontainers.CustomizeRequestOption {
	abs, err := filepath.Abs(scriptsDir)
	if err != nil {
		return func(req *testcontainers.GenericContainerRequest) error {
			return fmt.Errorf("get absolute path of scripts directory: %w", err)
		}
	}

	return func(req *testcontainers.GenericContainerRequest) error {
		prev := req.HostConfigModifier
		req.HostConfigModifier = func(hostConfig *container.HostConfig) {
			if prev != nil {
				prev(hostConfig)
			}

			// Remove any old /docker-entrypoint-initdb.d bind mounts.
			filtered := hostConfig.Binds[:0]
			for _, bind := range hostConfig.Binds {
				if !strings.HasSuffix(bind, ":/docker-entrypoint-initdb.d:ro") {
					filtered = append(filtered, bind)
				}
			}
			hostConfig.Binds = filtered

			// Mount the new scriptsDir
			hostConfig.Binds = append(hostConfig.Binds,
				fmt.Sprintf("%s:/docker-entrypoint-initdb.d:ro", abs))
		}

		return nil
	}
}

// WithMongotLogFile sets the path to the file where you want to store the logs
// of Atlas Search (mongot) by setting the MONGOT_LOG_FILE environment variable.
// Note that this can be set to /dev/stdout or /dev/stderr for convenience.
func WithMongotLogFile(logFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MONGOT_LOG_FILE"] = logFile

		return nil
	}
}

// WithRunnerLogFile sets the path to the file where you want to store the logs
// of "runner" but setting RUNNER_LOG_FILE environment variable. Note that this
// can be set to /dev/stdout or /dev/stderr for convenience.
func WithRunnerLogFile(logFile string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["RUNNER_LOG_FILE"] = logFile

		return nil
	}
}
