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
				// checking every 50ms, up to 1s, and failing if all the
				// exposed ports are not mapped in 5s.
				dockerContainer := c.(*DockerContainer)

				b := backoff.NewExponentialBackOff()

				b.InitialInterval = 50 * time.Millisecond
				b.MaxElapsedTime = 5 * time.Second
				b.MaxInterval = time.Duration(float64(time.Second) * backoff.DefaultRandomizationFactor)

				err := backoff.RetryNotify(
					func() error {
						jsonRaw, err := dockerContainer.inspectRawContainer(ctx)
						if err != nil {
							return err
						}

						return checkPortsMapped(jsonRaw.NetworkSettings.Ports, dockerContainer.exposedPorts)
					},
					b,
					func(err error, duration time.Duration) {
						dockerContainer.logger.Printf("All requested ports were not exposed: %v", err)
					},
				)
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
						return fmt.Errorf("wait until ready: %w", err)
					}
				}
				dockerContainer.isRunning = true
				return nil
			},
		},
	}
}

func checkPortsMapped(exposedAndMappedPorts nat.PortMap, exposedPorts []string) error {
	portMap, _, err := nat.ParsePortSpecs(exposedPorts)
	if err != nil {
		return fmt.Errorf("parse exposed ports: %w", err)
	}

	for exposedPort := range portMap {
		// having entries in exposedAndMappedPorts, where the key is the exposed port,
		// and the value is the mapped port, means that the port has been already mapped.
		if _, ok := exposedAndMappedPorts[exposedPort]; ok {
			continue
		}

		// check if the port is mapped with the protocol (default is TCP)
		if strings.Contains(string(exposedPort), "/") {
			return fmt.Errorf("port %s is not mapped yet", exposedPort)
		}

		// Port didn't have a type, default to tcp and retry.
		exposedPort += "/tcp"
		if _, ok := exposedAndMappedPorts[exposedPort]; !ok {
			return fmt.Errorf("port %s is not mapped yet", exposedPort)
		}
	}

	return nil
}
