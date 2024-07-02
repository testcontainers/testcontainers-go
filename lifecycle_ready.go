package testcontainers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/go-connections/nat"
)

// defaultReadinessHook is a hook that will wait for the container to be ready
var defaultReadinessHook = func() LifecycleHooks {
	return LifecycleHooks{
		PostStarts: []StartedContainerHook{
			func(ctx context.Context, c StartedContainer) error {
				// wait until all the exposed ports are mapped:
				// it will be ready when all the exposed ports are mapped,
				// checking every 50ms, up to 5s, and failing if all the
				// exposed ports are not mapped in that time.
				dockerContainer := c.(*DockerContainer)

				b := backoff.NewExponentialBackOff()

				b.InitialInterval = 50 * time.Millisecond
				b.MaxElapsedTime = 1 * time.Second
				b.MaxInterval = 5 * time.Second

				err := backoff.Retry(func() error {
					jsonRaw, err := dockerContainer.inspectRawContainer(ctx)
					if err != nil {
						return err
					}

					exposedAndMappedPorts := jsonRaw.NetworkSettings.Ports

					for _, exposedPort := range dockerContainer.exposedPorts {
						portMap := nat.Port(exposedPort)
						// having entries in exposedAndMappedPorts, where the key is the exposed port,
						// and the value is the mapped port, means that the port has been already mapped.
						if _, ok := exposedAndMappedPorts[portMap]; !ok {
							// check if the port is mapped with the protocol (default is TCP)
							if !strings.Contains(exposedPort, "/") {
								portMap = nat.Port(fmt.Sprintf("%s/tcp", exposedPort))
								if _, ok := exposedAndMappedPorts[portMap]; !ok {
									return fmt.Errorf("port %s is not mapped yet", exposedPort)
								}
							} else {
								return fmt.Errorf("port %s is not mapped yet", exposedPort)
							}
						}
					}

					return nil
				}, b)
				if err != nil {
					return fmt.Errorf("all exposed ports, %s, were not mapped in 5s: %w", dockerContainer.exposedPorts, err)
				}

				return nil
			},
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
