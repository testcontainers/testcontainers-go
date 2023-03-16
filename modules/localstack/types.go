package localstack

import (
	"github.com/testcontainers/testcontainers-go"
)

// Container represents the LocalStack container type used in the module
type Container struct {
	testcontainers.Container
}

// Deprecated: use Container instead
type LocalStackContainer struct {
	Container
}

// ContainerRequest represents the LocalStack container request type used in the module
// to configure the container
type ContainerRequest struct {
	testcontainers.ContainerRequest
}

// Deprecated: use ContainerRequest instead
type LocalStackContainerRequest struct {
	ContainerRequest
}

// OverrideContainerRequestOption is a type that can be used to configure the Testcontainers container request.
// The passed request will be merged with the default one.
// Deprecated: use testcontainers.CustomizeContainerRequestOption instead
type OverrideContainerRequestOption func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest

// NoopOverrideContainerRequest returns a helper function that does not override the container request
// Deprecated
var NoopOverrideContainerRequest = func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
	return req
}

// OverrideContainerRequest returns a function that can be used to merge the passed container request with one that is created by the LocalStack container
// Deprecated: use testcontainers.CustomizeContainerRequestOption instead
func OverrideContainerRequest(r testcontainers.ContainerRequest) func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
	return testcontainers.CustomizeContainerRequest(r)
}
