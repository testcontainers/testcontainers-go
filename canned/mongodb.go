package canned

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultUser     = ""
	defaultPassword = ""
	defaultDatabase = "database"
	defaultImage    = "mongo"
	defaultTag      = "4.2.0"
)

// MongoDbContainerRequest represents the parameters for requesting a running MongoDb container
type MongoDbContainerRequest struct {
	testcontainers.GenericContainerRequest
	User     string
	Password string
	Database string
}

type mongodbContainer struct {
	Container testcontainers.Container
	client    *mongo.Client
	req       MongoDbContainerRequest
	ctx       context.Context
}

func (c *mongodbContainer) GetDriver() (*mongo.Client, error) {

	host, err := c.Container.Host(c.ctx)
	if err != nil {
		return nil, err
	}

	port, err := c.Container.MappedPort(c.ctx, "27017/tcp")
	if err != nil {
		return nil, err
	}

	if c.req.User == "" {
		c.req.Password = defaultUser
	}

	if c.req.Password == "" {
		c.req.Password = defaultPassword
	}

	clientOptions := options.Client().ApplyURI(fmt.Sprintf(
		"mongodb://%s:%s@%s:%d",
		c.req.User,
		c.req.Password,
		host,
		port.Int(),
	))

	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// MongoDbContainer represents the running container instance.
func MongoDbContainer(ctx context.Context, req MongoDbContainerRequest) (*mongodbContainer, error) {

	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}

	req.ExposedPorts = []string{"27017/tcp"}
	req.Env = map[string]string{}
	req.Started = true

	if req.Image == "" {
		req.Image = fmt.Sprintf("%s:%s", defaultImage, defaultTag)
	}

	if req.User == "" {
		req.User = defaultUser
	}

	if req.Password == "" {
		req.Password = defaultPassword
	}

	if req.Database == "" {
		req.Database = defaultDatabase
	}

	req.Env["MONGO_INITDB_ROOT_USERNAME"] = req.User
	req.Env["MONGO_INITDB_ROOT_PASSWORD"] = req.Password
	req.Env["MONGO_INITDB_DATABASE"] = req.Database

	req.WaitingFor = wait.ForLog("waiting for connections on port 27017")

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	mongoDbC := &mongodbContainer{
		Container: c,
		req:       req,
		ctx:       ctx,
	}

	if req.Started {
		if err := c.Start(ctx); err != nil {
			return mongoDbC, errors.Wrap(err, "failed to start container")
		}
	}

	return mongoDbC, nil

}
