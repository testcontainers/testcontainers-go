package socat

import (
	"errors"
	"fmt"

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
type Option func(*options) error

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

// ExposedPort returns the exposed port of the target.
func (t Target) ExposedPort() int {
	return t.exposedPort
}

func (t Target) toCmd() string {
	return fmt.Sprintf("socat TCP-LISTEN:%d,fork,reuseaddr TCP:%s:%d", t.exposedPort, t.host, t.internalPort)
}

// NewTarget creates a new target for the socat container.
// The host of the target must be without the port,
// as it is internally mapped to the exposed port.
// The exposed port is exposed by the socat container.
func NewTarget(exposedPort int, host string) Target {
	return NewTargetWithInternalPort(exposedPort, exposedPort, host)
}

// NewTargetWithInternalPort creates a new target for the socat container.
// The host of the target must be without the port,
// as it is internally mapped to the exposed port.
// The exposed port is the port of the socat container, and
// the internal port is the port of the target container.
func NewTargetWithInternalPort(exposedPort int, internalPort int, host string) Target {
	// If the internal port is not set, use the exposed port
	if internalPort == 0 {
		internalPort = exposedPort
	}

	return Target{
		exposedPort:  exposedPort,
		internalPort: internalPort,
		host:         host,
	}
}

// WithTarget sets a single target for the socat container.
// The host of the target must be without the port, as it is internally mapped to the exposed port.
// Multiple calls to WithTarget will accumulate targets.
func WithTarget(target Target) Option {
	return func(o *options) error {
		if target.exposedPort == 0 {
			return errors.New("exposed port cannot be 0")
		}

		o.targets[target.exposedPort] = target

		newCmd := target.toCmd()

		if o.targetsCmd == "" {
			o.targetsCmd = newCmd
		} else {
			o.targetsCmd += " & " + newCmd
		}

		return nil
	}
}
