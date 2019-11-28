package postgres

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	image             = "postgres"
	tag               = "9.6.15"
	port     nat.Port = "5432/tcp"
	user              = "user"
	password          = "password"
	database          = "database"
	logEntry          = "database system is ready to accept connections"
)

type ContainerRequest struct {
	testcontainers.GenericContainerRequest
	Version  string
	User     string
	Password string
	Database string
}

// should always be created via NewContainer
type Container struct {
	Container testcontainers.Container
	req       ContainerRequest
}

func (c Container) ConnectURL(ctx context.Context) (string, error) {
	template := "postgres://%s:%s@%s:%d/%s"

	host, err := c.Container.Host(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get host")
	}

	mappedPort, err := c.Container.MappedPort(ctx, port)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get mapped port for %s", port.Port())
	}

	return fmt.Sprintf(template, c.req.User, c.req.Password, host,
		mappedPort.Int(), c.req.Database), nil
}

func NewContainer(ctx context.Context, req ContainerRequest) (*Container, error) {
	req.ExposedPorts = []string{string(port)}

	if req.Version != "" && req.FromDockerfile.Context == "" {
		req.Image = fmt.Sprintf("%s:%s", image, req.Version)
	}

	// Set the default values if none were provided in the request
	if req.Image == "" && req.FromDockerfile.Context == "" {
		req.Image = fmt.Sprintf("%s:%s", image, tag)
	}

	if req.User == "" {
		req.User = user
	}

	if req.Password == "" {
		req.Password = password
	}

	if req.Database == "" {
		req.Database = database
	}

	if req.Env == nil {
		req.Env = map[string]string{}
	}
	req.Env["POSTGRES_USER"] = req.User
	req.Env["POSTGRES_PASSWORD"] = req.Password
	req.Env["POSTGRES_DB"] = req.Database

	if req.WaitingFor == nil {
		req.WaitingFor = wait.ForAll(
			wait.NewHostPortStrategy(port),
			wait.ForLog(logEntry).WithOccurrence(2),
		)
	}

	if req.Cmd == nil {
		req.Cmd = []string{"postgres", "-c", "fsync=off"}
	}

	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	res := &Container{
		Container: c,
		req:       req,
	}

	if err := c.Start(ctx); err != nil {
		return res, errors.Wrap(err, "failed to start container")
	}

	return res, nil
}
