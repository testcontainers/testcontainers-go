package localstack

import (
	"github.com/testcontainers/testcontainers-go"
)

// LocalStackContainer represents the LocalStack container type used in the module
type LocalStackContainer struct {
	testcontainers.Container
}

// LocalStackContainerRequest represents the LocalStack container request type used in the module
// to configure the container
type LocalStackContainerRequest struct {
	testcontainers.GenericContainerRequest
}

// OverrideContainerRequestOption is a type that can be used to configure the Testcontainers container request.
// The passed request will be merged with the default one.
// Deprecated: use testcontainers.ContainerCustomizer instead
type OverrideContainerRequestOption func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest

// NoopOverrideContainerRequest returns a helper function that does not override the container request
// Deprecated: use testcontainers.ContainerCustomizer instead
var NoopOverrideContainerRequest = func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
	return req
}

func (opt OverrideContainerRequestOption) Customize(req *testcontainers.GenericContainerRequest) {
	req.ContainerRequest = opt(req.ContainerRequest)
}

// OverrideContainerRequest returns a function that can be used to merge the passed container request with one that is created by the LocalStack container
// Deprecated: use testcontainers.CustomizeRequest instead
func OverrideContainerRequest(r testcontainers.ContainerRequest) func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
	destContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: r,
	}

	return func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest {
		srcContainerReq := testcontainers.GenericContainerRequest{
			ContainerRequest: req,
		}

		opt := testcontainers.CustomizeRequest(destContainerReq)
		opt.Customize(&srcContainerReq)

		return srcContainerReq.ContainerRequest
	}
}
