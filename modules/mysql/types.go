package mysql

import "github.com/testcontainers/testcontainers-go"

type MySQLContainerRequest struct {
	testcontainers.ContainerRequest
}

type OverrideContainerRequestOption func(req testcontainers.ContainerRequest) testcontainers.ContainerRequest
