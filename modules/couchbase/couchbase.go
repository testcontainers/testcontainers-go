package couchbase

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
)

const (
	MGMT_PORT     = "8091"
	MGMT_SSL_PORT = "18091"

	VIEW_PORT     = "8092"
	VIEW_SSL_PORT = "18092"

	QUERY_PORT     = "8093"
	QUERY_SSL_PORT = "18093"

	SEARCH_PORT     = "8094"
	SEARCH_SSL_PORT = "18094"

	ANALYTICS_PORT     = "8095"
	ANALYTICS_SSL_PORT = "18095"

	EVENTING_PORT     = "8096"
	EVENTING_SSL_PORT = "18096"

	KV_PORT     = "11210"
	KV_SSL_PORT = "11207"
)

// CouchbaseContainer represents the Couchbase container type used in the module
type CouchbaseContainer struct {
	testcontainers.Container
}

// StartContainer creates an instance of the Couchbase container type
func StartContainer(ctx context.Context, opts ...Option) (*CouchbaseContainer, error) {
	config := &Config{
		enabledServices: []service{kv, query, search, index},
	}

	for _, opt := range opts {
		opt(config)
	}

	req := testcontainers.ContainerRequest{
		Image: "couchbase:6.5.1",
	}

	exposePorts(&req, config.enabledServices)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &CouchbaseContainer{Container: container}, nil
}

func exposePorts(req *testcontainers.ContainerRequest, enabledServices []service) {
	req.ExposedPorts = append(req.ExposedPorts, MGMT_PORT, MGMT_SSL_PORT)

	for _, service := range enabledServices {
		req.ExposedPorts = append(req.ExposedPorts, service.ports...)
	}
}
