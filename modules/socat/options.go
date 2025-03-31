package socat

import (
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	// targets is the map of targets of the socat container
	targets    map[int]Target
	targetsCmd string
}

func defaultOptions() options {
	return options{
		targets: map[int]Target{},
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Redpanda container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// Target represents a target for the socat container.
// Create a new target with NewTarget or NewTargetWithInternalPort.
type Target struct {
	exposedPort  int
	internalPort int
	host         string
}

func NewTarget(exposedPort int, host string) Target {
	return Target{
		exposedPort: exposedPort,
		host:        host,
	}
}

func NewTargetWithInternalPort(exposedPort int, internalPort int, host string) Target {
	return Target{
		exposedPort:  exposedPort,
		internalPort: internalPort,
		host:         host,
	}
}

// WithTargets sets the targets for the socat container.
// The host of each target must be without the port, as it is internally mapped to the exposed port.
func WithTargets(targets ...Target) Option {
	return func(o *options) {
		cmds := make([]string, 0, len(targets))

		// If the internal port is not set, use the exposed port
		for _, target := range targets {
			if target.internalPort == 0 {
				target.internalPort = target.exposedPort
			}

			o.targets[target.exposedPort] = target

			cmds = append(cmds, fmt.Sprintf("socat TCP-LISTEN:%d,fork,reuseaddr TCP:%s:%d", target.exposedPort, target.host, target.internalPort))
		}

		o.targetsCmd = strings.Join(cmds, " & ")
	}
}
