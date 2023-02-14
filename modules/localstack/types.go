package localstack

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/testcontainers/testcontainers-go"
)

// LocalStackContainer represents the LocalStack container type used in the module
type LocalStackContainer struct {
	testcontainers.Container
}

// LocalStackContainerRequest represents the LocalStack container request type used in the module
// to configure the container
type LocalStackContainerRequest struct {
	testcontainers.ContainerRequest
}

// OverrideContainerRequestOption is a type that can be used to configure the Testcontainers container request.
// The passed request will be merged with the default one.
type OverrideContainerRequestOption func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest

// NoopOverrideContainerRequest returns a helper function that does not override the container request
var NoopOverrideContainerRequest = func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
	return req
}

// OverrideContainerRequest returns a function that can be used to merge the passed container request with one that is created by the LocalStack container
func OverrideContainerRequest(r testcontainers.ContainerRequest) func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
	return func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
		if err := mergo.Merge(&req, r, mergo.WithOverride); err != nil {
			fmt.Printf("error merging container request %v. Keeping the default one: %v", err, req)
			return req
		}

		return req
	}
}
