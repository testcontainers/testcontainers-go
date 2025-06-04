package mongodb

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed mount/entrypoint-tc.sh
var entrypointContent []byte

const (
	entrypointPath      = "/tmp/entrypoint-tc.sh"
	keyFilePath         = "/tmp/mongo_keyfile"
	replicaSetOptEnvKey = "testcontainers.mongodb.replicaset_name"
)

// MongoDBContainer represents the MongoDB container type used in the module
type MongoDBContainer struct {
	testcontainers.Container
	username   string
	password   string
	replicaSet string
}

// Deprecated: use Run instead
// RunContainer creates an instance of the MongoDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MongoDBContainer, error) {
	return Run(ctx, "mongo:6", opts...)
}

// Run creates an instance of the MongoDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MongoDBContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("27017/tcp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("Waiting for connections"),
			wait.ForListeningPort("27017/tcp"),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	defaultOptions := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&defaultOptions); err != nil {
				return nil, fmt.Errorf("mongodb option: %w", err)
			}
		}
	}

	username := defaultOptions.env["MONGO_INITDB_ROOT_USERNAME"]
	password := defaultOptions.env["MONGO_INITDB_ROOT_PASSWORD"]
	if username != "" && password == "" || username == "" && password != "" {
		return nil, errors.New("if you specify username or password, you must provide both of them")
	}

	replicaSet := defaultOptions.env[replicaSetOptEnvKey]
	if replicaSet != "" {
		if username == "" || password == "" {
			moduleOpts = append(moduleOpts, noAuthReplicaSet(replicaSet))
		} else {
			moduleOpts = append(moduleOpts, withAuthReplicaset(replicaSet, username, password))
		}
	}

	moduleOpts = append(moduleOpts, testcontainers.WithEnv(defaultOptions.env))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *MongoDBContainer
	if ctr != nil {
		c = &MongoDBContainer{Container: ctr, username: username, password: password, replicaSet: replicaSet}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}

// ConnectionString returns the connection string for the MongoDB container.
// If you provide a username and a password, the connection string will also include them.
func (c *MongoDBContainer) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := c.MappedPort(ctx, "27017/tcp")
	if err != nil {
		return "", err
	}
	u := url.URL{
		Scheme: "mongodb",
		Host:   net.JoinHostPort(host, port.Port()),
		Path:   "/",
	}

	if c.username != "" && c.password != "" {
		u.User = url.UserPassword(c.username, c.password)
	}

	if c.replicaSet != "" {
		q := url.Values{}
		q.Add("replicaSet", c.replicaSet)
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

func setupEntrypointForAuth() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if err := testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader:            bytes.NewReader(entrypointContent),
			ContainerFilePath: entrypointPath,
			FileMode:          0o755,
		})(req); err != nil {
			return err
		}

		if err := testcontainers.WithEntrypoint(entrypointPath)(req); err != nil {
			return err
		}

		if err := testcontainers.WithEnv(map[string]string{
			"MONGO_KEYFILE": keyFilePath,
		})(req); err != nil {
			return err
		}

		return nil
	}
}

func noAuthReplicaSet(replSetName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cli := newMongoCli("", "")
		if err := testcontainers.WithCmdArgs("--replSet", replSetName)(req); err != nil {
			return fmt.Errorf("with cmd args: %w", err)
		}

		if err := initiateReplicaSet(cli, replSetName)(req); err != nil {
			return fmt.Errorf("initiate replica set: %w", err)
		}

		return nil
	}
}

func initiateReplicaSet(cli mongoCli, replSetName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.WaitingFor = wait.ForAll(
			req.WaitingFor,
			wait.ForExec(cli.eval("rs.status().ok")),
		).WithDeadline(60 * time.Second)

		req.LifecycleHooks = append(
			req.LifecycleHooks, testcontainers.ContainerLifecycleHooks{
				PostStarts: []testcontainers.ContainerHook{
					func(ctx context.Context, c testcontainers.Container) error {
						ip, err := c.ContainerIP(ctx)
						if err != nil {
							return fmt.Errorf("container ip: %w", err)
						}

						cmd := cli.eval(
							"rs.initiate({ _id: '%s', members: [ { _id: 0, host: '%s:27017' } ] })",
							replSetName,
							ip,
						)
						return wait.ForExec(cmd).WaitUntilReady(ctx, c)
					},
				},
			},
		)

		return nil
	}
}

func withAuthReplicaset(
	replSetName string,
	username string,
	password string,
) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if err := setupEntrypointForAuth()(req); err != nil {
			return fmt.Errorf("setup entrypoint for auth: %w", err)
		}

		cli := newMongoCli(username, password)

		if err := testcontainers.WithCmdArgs("--replSet", replSetName, "--keyFile", keyFilePath)(req); err != nil {
			return fmt.Errorf("with cmd args: %w", err)
		}

		if err := initiateReplicaSet(cli, replSetName)(req); err != nil {
			return fmt.Errorf("initiate replica set: %w", err)
		}

		return nil
	}
}
