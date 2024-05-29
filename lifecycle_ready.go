package testcontainers

import (
	"context"
)

// defaultReadinessHook is a hook that will wait for the container to be ready
var defaultReadinessHook = func() ContainerLifecycleHooks {
	return ContainerLifecycleHooks{
		PostStarts: []StartedContainerHook{
			// wait for the container to be ready
			func(ctx context.Context, c StartedContainer) error {
				return c.WaitUntilReady(ctx)
			},
		},
	}
}
