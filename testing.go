package testcontainers

import (
	"context"
	"testing"
)

// SkipIfProviderIsNotHealthy is a utility function capable of skipping tests
// if the provider is not healthy, or running at all.
// This is a function designed to be used in your test, when Docker is not mandatory for CI/CD.
// In this way tests that depend on Testcontainers won't run if the provider is provisioned correctly.
func SkipIfProviderIsNotHealthy(t *testing.T) {
	ctx := context.Background()
	provider, err := ProviderDocker.GetProvider()
	if err != nil {
		t.Skipf("Docker is not running. TestContainers can't perform is work without it: %s", err)
	}
	err = provider.Health(ctx)
	if err != nil {
		t.Skipf("Docker is not running. TestContainers can't perform is work without it: %s", err)
	}
}

// Cleanup is a utility function used to cleanup containers in a testing
// context.
func Cleanup(tb testing.TB, ctx context.Context, ctr Container) {
	tb.Helper()
	tb.Cleanup(func() {
		tb.Logf("terminating container \"%s\"\n", ctr.GetContainerID())
		if err := ctr.Terminate(ctx); err != nil {
			tb.Fatalf("failed to terminate container \":%s\": %s\n", ctr.GetContainerID(), err)
		}
	})
}
