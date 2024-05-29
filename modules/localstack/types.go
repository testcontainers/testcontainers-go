package localstack

import (
	"github.com/testcontainers/testcontainers-go"
)

// Container represents the LocalStack container type used in the module
type Container struct {
	*testcontainers.DockerContainer
}
