package mongodb

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed mount/entrypoint-tc.sh
var entrypointContent []byte

const (
	defaultPort         = "27017/tcp"
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
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Waiting for connections"),
			wait.ForListeningPort(defaultPort),
		),
		testcontainers.WithEnv(map[string]string{}),
	}

	moduleOpts = append(moduleOpts, opts...)

	// configure the request for the replicaset after all the options have been applied
	moduleOpts = append(moduleOpts, configureReplicaset())

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *MongoDBContainer
	if container != nil {
		c = &MongoDBContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("run mongodb: %w", err)
	}

	// Inspect the container to get environment variables
	inspect, err := container.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect mongodb: %w", err)
	}

	// refresh the credentials from the environment variables
	foundUser, foundPass, foundReplicaSet := false, false, false
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "MONGO_INITDB_ROOT_USERNAME="); ok {
			c.username, foundUser = v, true
		}
		if v, ok := strings.CutPrefix(env, "MONGO_INITDB_ROOT_PASSWORD="); ok {
			c.password, foundPass = v, true
		}
		if v, ok := strings.CutPrefix(env, replicaSetOptEnvKey+"="); ok {
			c.replicaSet, foundReplicaSet = v, true
		}

		if foundUser && foundPass && foundReplicaSet {
			break
		}
	}

	return c, nil
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a username and its password.
// It will create the specified user with superuser power.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"MONGO_INITDB_ROOT_USERNAME": username,
	})
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is used in conjunction with WithUsername to set a username and its password.
// It will set the superuser password for MongoDB.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"MONGO_INITDB_ROOT_PASSWORD": password,
	})
}

// WithReplicaSet sets the replica set name for Single node MongoDB replica set.
func WithReplicaSet(replSetName string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		replicaSetOptEnvKey: replSetName,
	})
}

// ConnectionString returns the connection string for the MongoDB container.
// If you provide a username and a password, the connection string will also include them.
func (c *MongoDBContainer) ConnectionString(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, defaultPort, "")
	if err != nil {
		return "", err
	}
	u := url.URL{
		Scheme: "mongodb",
		Host:   endpoint,
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
			return fmt.Errorf("files: %w", err)
		}

		if err := testcontainers.WithEntrypoint(entrypointPath)(req); err != nil {
			return fmt.Errorf("entrypoint: %w", err)
		}

		if err := testcontainers.WithEnv(map[string]string{
			"MONGO_KEYFILE": keyFilePath,
		})(req); err != nil {
			return fmt.Errorf("env: %w", err)
		}

		return nil
	}
}

func configureReplicaset() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		username := req.Env["MONGO_INITDB_ROOT_USERNAME"]
		password := req.Env["MONGO_INITDB_ROOT_PASSWORD"]
		if username != "" && password == "" || username == "" && password != "" {
			return errors.New("if you specify username or password, you must provide both of them")
		}

		replicaSet := req.Env[replicaSetOptEnvKey]
		if replicaSet != "" {
			if username == "" || password == "" {
				return noAuthReplicaSet(replicaSet)(req)
			}

			return withAuthReplicaset(replicaSet, username, password)(req)
		}

		return nil
	}
}

func noAuthReplicaSet(replSetName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cli := newMongoCli("", "")

		if err := testcontainers.WithCmdArgs("--replSet", replSetName)(req); err != nil {
			return fmt.Errorf("cmd args: %w", err)
		}

		return initiateReplicaSet(cli, replSetName)(req)
	}
}

func initiateReplicaSet(cli mongoCli, replSetName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.WaitingFor = wait.ForAll(
			req.WaitingFor,
			wait.ForExec(cli.eval("rs.status().ok")),
		).WithDeadline(60 * time.Second)

		if err := testcontainers.WithAdditionalLifecycleHooks(
			testcontainers.ContainerLifecycleHooks{
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
		)(req); err != nil {
			return fmt.Errorf("lifecycle hooks: %w", err)
		}

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

		if err := testcontainers.WithCmdArgs(
			"--replSet", replSetName,
			"--keyFile", keyFilePath,
		)(req); err != nil {
			return fmt.Errorf("cmd args: %w", err)
		}

		return initiateReplicaSet(cli, replSetName)(req)
	}
}
