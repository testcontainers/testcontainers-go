// This test is testing very internal logic that should not be exported away from this package. We'll
// leave it in the main testcontainers package. Do not use for user facing examples.
package testcontainers

import "testing"

func ExampleSkipIfProviderIsNotHealthy() {
	SkipIfProviderIsNotHealthy(&testing.T{})
}
