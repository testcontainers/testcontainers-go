package testcontainers

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
)

// WaitForMappedPort waits for given local port to be mapped by checking mapping
// until specified context is cancelled with given interval b/w checks.
// Returns externally mapped port or error.
func WaitForMappedPort(ctx context.Context, c Container, localPort nat.Port, interval time.Duration) (nat.Port, error) {
	mappedPort, err := c.MappedPort(ctx, localPort)
	for i := 1; err != nil; i++ {
		select {
		case <-ctx.Done():
			return mappedPort, fmt.Errorf(
				"mapped port: retries: %d, local port: %s, last err: %w, ctx err: %w", i, localPort, err, ctx.Err())
		case <-time.After(interval):
			mappedPort, err = c.MappedPort(ctx, localPort)
		}
	}
	return mappedPort, nil
}
