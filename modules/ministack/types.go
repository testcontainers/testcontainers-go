package ministack

import (
	"github.com/testcontainers/testcontainers-go"
)

// Container represents the ministack container type used in the module
type Container struct {
	testcontainers.Container
}
