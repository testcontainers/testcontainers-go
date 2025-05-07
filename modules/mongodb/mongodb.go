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
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Waiting for connections"),
			wait.ForListeningPort("27017/tcp"),
		),
		Env: map[string]string{},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}
	username := req.Env["MONGO_INITDB_ROOT_USERNAME"]
	password := req.Env["MONGO_INITDB_ROOT_PASSWORD"]
	if username != "" && password == "" || username == "" && password != "" {
		return nil, errors.New("if you specify username or password, you must provide both of them")
	}

	replicaSet := req.Env[replicaSetOptEnvKey]
	if replicaSet != "" {
		if err := configureRequestForReplicaset(username, password, replicaSet, &genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *MongoDBContainer
	if container != nil {
		c = &MongoDBContainer{Container: container, username: username, password: password, replicaSet: replicaSet}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a username and its password.
// It will create the specified user with superuser power.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MONGO_INITDB_ROOT_USERNAME"] = username

		return nil
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is used in conjunction with WithUsername to set a username and its password.
// It will set the superuser password for MongoDB.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MONGO_INITDB_ROOT_PASSWORD"] = password

		return nil
	}
}

// WithReplicaSet sets the replica set name for Single node MongoDB replica set.
func WithReplicaSet(replSetName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[replicaSetOptEnvKey] = replSetName

		return nil
	}
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

func setupEntrypointForAuth(req *testcontainers.GenericContainerRequest) {
	req.Files = append(
		req.Files, testcontainers.ContainerFile{
			Reader:            bytes.NewReader(entrypointContent),
			ContainerFilePath: entrypointPath,
			FileMode:          0o755,
		},
	)
	req.Entrypoint = []string{entrypointPath}
	req.Env["MONGO_KEYFILE"] = keyFilePath
}

func configureRequestForReplicaset(
	username string,
	password string,
	replicaSet string,
	genericContainerReq *testcontainers.GenericContainerRequest,
) error {
	if username == "" || password == "" {
		return noAuthReplicaSet(replicaSet)(genericContainerReq)
	}

	return withAuthReplicaset(replicaSet, username, password)(genericContainerReq)
}

func noAuthReplicaSet(replSetName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cli := newMongoCli("", "")
		req.Cmd = append(req.Cmd, "--replSet", replSetName)
		initiateReplicaSet(req, cli, replSetName)

		return nil
	}
}

func initiateReplicaSet(req *testcontainers.GenericContainerRequest, cli mongoCli, replSetName string) {
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
}

func withAuthReplicaset(
	replSetName string,
	username string,
	password string,
) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		setupEntrypointForAuth(req)
		cli := newMongoCli(username, password)
		req.Cmd = append(req.Cmd, "--replSet", replSetName, "--keyFile", keyFilePath)
		initiateReplicaSet(req, cli, replSetName)

		return nil
	}
}
