package couchbase

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// CouchbaseContainer represents the Couchbase container type used in the module
type CouchbaseContainer struct {
	testcontainers.Container
}

// StartContainer creates an instance of the Couchbase container type
func StartContainer(ctx context.Context) (*CouchbaseContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "couchbase:6.5.1",
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &CouchbaseContainer{Container: container}, nil
}
