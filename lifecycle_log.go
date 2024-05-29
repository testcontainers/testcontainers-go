package testcontainers

import (
	"context"

	"github.com/testcontainers/testcontainers-go/log"
)

// DefaultLoggingHook is a hook that will log the container lifecycle events
var DefaultLoggingHook = func(logger log.Logging) ContainerLifecycleHooks {
	shortContainerID := func(c CreatedContainer) string {
		return c.GetContainerID()[:12]
	}

	return ContainerLifecycleHooks{
		PreCreates: []ContainerRequestHook{
			func(ctx context.Context, def ContainerDefinition) error {
				logger.Printf("🐳 Creating container for image %s", def.GetImage())
				return nil
			},
		},
		PostCreates: []CreatedContainerHook{
			func(ctx context.Context, c CreatedContainer) error {
				logger.Printf("✅ Container created: %s", shortContainerID(c))
				return nil
			},
		},
		PreStarts: []CreatedContainerHook{
			func(ctx context.Context, c CreatedContainer) error {
				logger.Printf("🐳 Starting container: %s", shortContainerID(c))
				return nil
			},
		},
		PostStarts: []StartedContainerHook{
			func(ctx context.Context, c StartedContainer) error {
				logger.Printf("✅ Container started: %s", shortContainerID(c))
				return nil
			},
		},
		PostReadies: []StartedContainerHook{
			func(ctx context.Context, c StartedContainer) error {
				logger.Printf("🔔 Container is ready: %s", shortContainerID(c))
				return nil
			},
		},
		PreStops: []StartedContainerHook{
			func(ctx context.Context, c StartedContainer) error {
				logger.Printf("🐳 Stopping container: %s", shortContainerID(c))
				return nil
			},
		},
		PostStops: []StartedContainerHook{
			func(ctx context.Context, c StartedContainer) error {
				logger.Printf("✅ Container stopped: %s", shortContainerID(c))
				return nil
			},
		},
		PreTerminates: []StartedContainerHook{
			func(ctx context.Context, c StartedContainer) error {
				logger.Printf("🐳 Terminating container: %s", shortContainerID(c))
				return nil
			},
		},
		PostTerminates: []StartedContainerHook{
			func(ctx context.Context, c StartedContainer) error {
				logger.Printf("🚫 Container terminated: %s", shortContainerID(c))
				return nil
			},
		},
	}
}

// defaultLogConsumersHook is a hook that will start log consumers after the container is started
var defaultLogConsumersHook = func(cfg *log.ConsumerConfig) ContainerLifecycleHooks {
	return ContainerLifecycleHooks{
		PostStarts: []StartedContainerHook{
			// first post-start hook is to produce logs and start log consumers
			func(ctx context.Context, c StartedContainer) error {
				if cfg == nil {
					return nil
				}

				for _, consumer := range cfg.Consumers {
					c.FollowOutput(consumer)
				}

				if len(cfg.Consumers) > 0 {
					return c.StartLogProduction(ctx, cfg.Opts...)
				}
				return nil
			},
		},
		PreTerminates: []StartedContainerHook{
			// first pre-terminate hook is to stop the log production
			func(ctx context.Context, c StartedContainer) error {
				if cfg == nil || len(cfg.Consumers) == 0 {
					return nil
				}

				return c.StopLogProduction()
			},
		},
	}
}
