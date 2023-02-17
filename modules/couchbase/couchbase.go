package couchbase

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
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

	if err := waitUntilAllNodesAreHealthy(ctx, container, config.enabledServices); err != nil {
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

func waitUntilAllNodesAreHealthy(ctx context.Context, container testcontainers.Container, enabledServices []service) error {
	var waitStrategy []wait.Strategy

	waitStrategy = append(waitStrategy, wait.ForHTTP("/pools/default").
		WithPort(MGMT_PORT).
		WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusOK
		}).
		WithResponseMatcher(func(body io.Reader) bool {
			json, err := io.ReadAll(body)
			if err != nil {
				return false
			}
			status := gjson.Get(string(json), "nodes.0.status")
			if status.String() != "healthy" {
				return false
			}

			return true
		}))

	for _, service := range enabledServices {
		var strategy wait.Strategy

		switch service.identifier {
		case query.identifier:
			strategy = wait.ForHTTP("/admin/ping").
				WithPort(QUERY_PORT).
				WithStatusCodeMatcher(func(status int) bool {
					return status == http.StatusOK
				})
		case analytics.identifier:
			strategy = wait.ForHTTP("/admin/ping").
				WithPort(ANALYTICS_PORT).
				WithStatusCodeMatcher(func(status int) bool {
					return status == http.StatusOK
				})
		case eventing.identifier:
			strategy = wait.ForHTTP("/api/v1/config").
				WithPort(EVENTING_PORT).
				WithStatusCodeMatcher(func(status int) bool {
					return status == http.StatusOK
				})
		}

		if strategy != nil {
			waitStrategy = append(waitStrategy, strategy)
		}
	}

	return wait.ForAll(waitStrategy...).WaitUntilReady(ctx, container)
}
