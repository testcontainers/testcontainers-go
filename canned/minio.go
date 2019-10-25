package canned

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	minio "github.com/minio/minio-go/v6"
)

const (
	minioAccessKey  = "AKIAIOSFODNN7EXAMPLE"
	minioSecretKey  = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	minioImage      = "minio/minio"
	minioDefaultTag = "latest"
	minioPort       = "9000/tcp"
)

// MinioContainerRequest adds some Minio specific parameters
// to GenericContainerRequest
type MinioContainerRequest struct {
	testcontainers.GenericContainerRequest
	AccessKey string
	SecretKey string
}

// MinioContainer should always be created via NewMinioContainer
type MinioContainer struct {
	Container testcontainers.Container
	client    *minio.Client
	req       MinioContainerRequest
}

// GetClient provides a Minio Go client as described here
// https://docs.min.io/docs/golang-client-api-reference.html
func (c *MinioContainer) GetClient(ctx context.Context) (*minio.Client, error) {

	host, err := c.Container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := c.Container.MappedPort(ctx, minioPort)
	if err != nil {
		return nil, err
	}

	client, err := minio.New(
		fmt.Sprintf("%s:%d", host, mappedPort.Int()),
		c.req.AccessKey,
		c.req.SecretKey,
		false,
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewMinioContainer creates and (optionally) starts a Minio (S3-compatible object storage) container.
// If autostarted, the function waits for port 9000/tcp to be listening,
// and the log to have shown the web access endpoints.
func NewMinioContainer(ctx context.Context, req MinioContainerRequest) (*MinioContainer, error) {

	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}

	// With the current logic it's not really possible to allow other ports...
	req.ExposedPorts = []string{minioPort}

	if req.Env == nil {
		req.Env = map[string]string{}
	}

	// Set the default values if none were provided in the request
	if req.Image == "" && req.FromDockerfile.Context == "" {
		req.Image = fmt.Sprintf("%s:%s", minioImage, minioDefaultTag)
	}

	if req.AccessKey == "" {
		req.AccessKey = minioAccessKey
	}

	if req.SecretKey == "" {
		req.SecretKey = minioSecretKey
	}

	req.Env["MINIO_ACCESS_KEY"] = req.AccessKey
	req.Env["MINIO_SECRET_KEY"] = req.SecretKey

	req.ContainerRequest.Cmd = []string{"server", "/data"}

	req.WaitingFor = wait.ForAll(
		wait.ForListeningPort(minioPort),
		wait.ForLog("Browser Access:"),
	)

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create container")
	}

	minioC := &MinioContainer{
		Container: c,
		req:       req,
	}

	if req.Started {
		if err := c.Start(ctx); err != nil {
			return minioC, errors.Wrap(err, "failed to start container")
		}
	}

	return minioC, nil
}
