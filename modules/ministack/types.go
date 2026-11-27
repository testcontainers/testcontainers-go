package ministack

import (
	"github.com/testcontainers/testcontainers-go"
)

// MinistackContainer represents the ministack container type used in the module
type MinistackContainer struct {
	testcontainers.Container
}
