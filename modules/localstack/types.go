package localstack

import (
	"github.com/testcontainers/testcontainers-go"
)

// LocalStackContainer represents the LocalStack container type used in the module
type LocalStackContainer struct {
	*testcontainers.DockerContainer
}
