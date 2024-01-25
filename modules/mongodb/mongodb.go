package mongodb

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// defaultImage is the default MongoDB container image
const defaultImage = "mongo:6"

// MongoDBContainer represents the MongoDB container type used in the module
type MongoDBContainer struct {
	testcontainers.Container
	username   string
	password   string
	replicaSet string
}

// RunContainer creates an instance of the MongoDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MongoDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultImage,
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
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}
	mongoContainer := &MongoDBContainer{Container: container}

	username := req.Env["MONGO_INITDB_ROOT_USERNAME"]
	password := req.Env["MONGO_INITDB_ROOT_PASSWORD"]
	if username != "" && password == "" || username == "" && password != "" {
		return nil, fmt.Errorf("if you specify username or password, you must provide both of them")
	}

	if username != "" && password != "" {
		mongoContainer.username = username
		mongoContainer.password = password
	}

	replicaSet := req.Env["MONGO_REPLICASET_NAME"]
	if replicaSet != "" {
		mongoContainer.replicaSet = replicaSet
	}

	return mongoContainer, nil
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a username and its password.
// It will create the specified user with superuser power.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MONGO_INITDB_ROOT_USERNAME"] = username
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is used in conjunction with WithUsername to set a username and its password.
// It will set the superuser password for MongoDB.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MONGO_INITDB_ROOT_PASSWORD"] = password
	}
}

// WithReplicaSet TODO: help me fill this func comment
func WithReplicaSet(rsName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["MONGO_REPLICASET_NAME"] = rsName
		req.Cmd = append(req.Cmd, "--replSet", rsName)
		req.LifecycleHooks = append(req.LifecycleHooks, replicaSetLifecycleHooks())
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

	connStr := strings.Builder{}
	connStr.WriteString("mongodb://")

	if c.username != "" && c.password != "" {
		connStr.WriteString(fmt.Sprintf("%s:%s@", c.username, c.password))
		return fmt.Sprintf("mongodb://%s:%s@%s:%s", c.username, c.password, host, port.Port()), nil
	}

	connStr.WriteString(fmt.Sprintf("%s:%s", host, port.Port()))

	if c.replicaSet != "" {
		connStr.WriteString(fmt.Sprintf("/test?replicaSet=%s&directConnection=true", c.replicaSet))
	}

	return connStr.String(), nil
}

func replicaSetLifecycleHooks() testcontainers.ContainerLifecycleHooks {
	return testcontainers.ContainerLifecycleHooks{
		PostStarts: []testcontainers.ContainerHook{
			replicaSetPostStart,
		},
	}
}

func replicaSetPostStart(ctx context.Context, container testcontainers.Container) error {
	const (
		waitScript = `var attempt = 0;
		while (%s) {
			if (attempt > %d) {
				quit(1);
			}

			print('%s ' + attempt);
			sleep(100);
			attempt++;
		}`
		waitCond        = "db.runCommand( { isMaster: 1 } ).ismaster==false"
		maxWaitAttempts = 60
	)

	exitCode, r, err := container.Exec(ctx, []string{"mongosh", "--eval", "rs.initiate();"})
	if err != nil {
		return err
	}
	if exitCode != 0 {
		dataVjp, _ := io.ReadAll(r)
		return fmt.Errorf("non-zero exit code to initiate replica: %d, %s", exitCode, string(dataVjp))
	}

	waitCmd := fmt.Sprintf(waitScript, waitCond, maxWaitAttempts, "An attempt to await for a single node replica set initialization:")
	exitCode, _, err = container.Exec(ctx, []string{"mongosh", "--eval", waitCmd})
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("none-zero exit code when await replica initiate")
	}

	return nil
}
