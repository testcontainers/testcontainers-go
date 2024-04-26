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

// Deprecated: use testcontainers.ContainerCustomizer instead
// OverrideContainerRequestOption is a type that can be used to configure the Testcontainers container request.
// The passed request will be merged with the default one.
type OverrideContainerRequestOption func(req testcontainers.ContainerRequest) (testcontainers.ContainerRequest, error)

// Deprecated: use testcontainers.ContainerCustomizer instead
// NoopOverrideContainerRequest returns a helper function that does not override the container request
var NoopOverrideContainerRequest = func(req testcontainers.ContainerRequest) (testcontainers.ContainerRequest, error) {
	return req, nil
}

// Deprecated: use testcontainers.ContainerCustomizer instead
func (opt OverrideContainerRequestOption) Customize(req *testcontainers.GenericContainerRequest) error {
	r, err := opt(req.ContainerRequest)
	if err != nil {
		return err
	}

	req.ContainerRequest = r

	return nil
}

// Deprecated: use testcontainers.CustomizeRequest instead
// OverrideContainerRequest returns a function that can be used to merge the passed container request with one that is created by the LocalStack container
func OverrideContainerRequest(r testcontainers.ContainerRequest) func(req testcontainers.ContainerRequest) (testcontainers.ContainerRequest, error) {
	destContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: r,
	}

	return func(req testcontainers.ContainerRequest) (testcontainers.ContainerRequest, error) {
		srcContainerReq := testcontainers.GenericContainerRequest{
			ContainerRequest: req,
		}

		opt := testcontainers.CustomizeRequest(destContainerReq)
		if err := opt.Customize(&srcContainerReq); err != nil {
			return testcontainers.ContainerRequest{}, err
		}

		return srcContainerReq.ContainerRequest, nil
	}
}
