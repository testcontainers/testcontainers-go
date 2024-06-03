package testcontainers

import (
	"context"
)

// defaultReadinessHook is a hook that will wait for the container to be ready
var defaultReadinessHook = func() LifecycleHooks {
	return LifecycleHooks{
		PostStarts: []StartedContainerHook{
			// wait for the container to be ready
			func(ctx context.Context, c StartedContainer) error {
				dockerContainer := c.(*DockerContainer)
				// if a Wait Strategy has been specified, wait before returning
				if dockerContainer.WaitingFor != nil {
					dockerContainer.logger.Printf(
						"⏳ Waiting for container id %s image: %s. Waiting for: %+v",
						dockerContainer.ID[:12], dockerContainer.Image, dockerContainer.WaitingFor,
					)
					if err := dockerContainer.WaitingFor.WaitUntilReady(ctx, c); err != nil {
						return err
					}
				}
				dockerContainer.isRunning = true
				return nil
			},
		},
	}
}
