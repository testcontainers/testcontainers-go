package atlaslocal

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const defaultPort = "27017/tcp"

// Container represents the MongoDBAtlasLocal container type used in the module.
type Container struct {
	testcontainers.Container
	userOpts options
}

// Run creates an instance of the MongoDBAtlasLocal container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	userOpts := options{}
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&userOpts); err != nil {
				return nil, fmt.Errorf("apply option: %w", err)
			}
		}
	}

	if err := userOpts.validate(); err != nil {
		return nil, fmt.Errorf("validate options: %w", err)
	}

	moduleOpts := []testcontainers.ContainerCustomizer{ // Set the defaults
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithWaitStrategy(wait.ForAll(wait.ForListeningPort(defaultPort), wait.ForHealthCheck())),
		testcontainers.WithEnv(userOpts.env()),
		testcontainers.WithFiles(userOpts.files...),
	}

	for _, opt := range opts {
		moduleOpts = append(moduleOpts, opt)
	}

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if container != nil {
		c = &Container{Container: container, userOpts: userOpts}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// ConnectionString returns the connection string for the MongoDB Atlas Local
// container. If you provide a username and a password, the connection string
// will also include them.
func (ctr *Container) ConnectionString(ctx context.Context) (string, error) {
	host, err := ctr.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := ctr.MappedPort(ctx, "27017/tcp")
	if err != nil {
		return "", err
	}

	uri := &url.URL{
		Scheme: "mongodb",
		Host:   net.JoinHostPort(host, mappedPort.Port()),
		Path:   "/",
	}

	// If MONGODB_INITDB_DATABASE is set, use it as the default database in the
	// connection string.
	if db := ctr.userOpts.database; db != "" {
		uri.Path, err = url.JoinPath("/", db)
		if err != nil {
			return "", fmt.Errorf("join path: %w", err)
		}
	}

	user, err := ctr.userOpts.parseUsername()
	if err != nil {
		return "", fmt.Errorf("parse username: %w", err)
	}

	password, err := ctr.userOpts.parsePassword()
	if err != nil {
		return "", fmt.Errorf("parse password: %w", err)
	}

	if user != "" && password != "" {
		uri.User = url.UserPassword(user, password)
	}

	q := uri.Query()
	q.Set("directConnection", "true")
	if user != "" && password != "" {
		q.Set("authSource", "admin")
	}

	uri.RawQuery = q.Encode()

	return uri.String(), nil
}

// ReadMongotLogs returns a reader for mongot logs in the container. Reads from
// stdout/stderr or /tmp/mongot.log if configured.
//
// This method return the os.ErrNotExist sentinel error if it is called with
// no log file configured.
func (ctr *Container) ReadMongotLogs(ctx context.Context) (io.ReadCloser, error) {
	path := ctr.userOpts.mongotLogPath
	if path == "" {
		return nil, os.ErrNotExist
	}

	switch path {
	case "/dev/stdout", "/dev/stderr":
		return ctr.Logs(ctx)
	default:
		return ctr.CopyFileFromContainer(ctx, path)
	}
}

// ReadRunnerLogs() returns a reader for runner logs in the container. Reads
// from stdout/stderr or /tmp/runner.log if configured.
//
// This method return the os.ErrNotExist sentinel error if it is called with
// no log file configured.
func (ctr *Container) ReadRunnerLogs(ctx context.Context) (io.ReadCloser, error) {
	path := ctr.userOpts.runnerLogPath
	if path == "" {
		return nil, os.ErrNotExist
	}

	switch path {
	case "/dev/stdout", "/dev/stderr":
		return ctr.Logs(ctx)
	default:
		return ctr.CopyFileFromContainer(ctx, path)
	}
}
