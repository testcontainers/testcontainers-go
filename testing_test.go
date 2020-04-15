package testcontainers

import "testing"

func ExampleSkipIfProviderIsNotHealthy() {
	SkipIfProviderIsNotHealthy(&testing.T{})
}
