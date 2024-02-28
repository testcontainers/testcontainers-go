package testcontainers_test

import (
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

func ExampleSkipIfProviderIsNotHealthy() {
	testcontainers.SkipIfProviderIsNotHealthy(&testing.T{})
}
