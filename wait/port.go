package wait

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
)

const (
	// unboundIPv4 is the address used when the port is not bound to a specific IP.
	unboundIPv4 = "0.0.0.0"
	// unboundIPv6 is the address used when the port is not bound to a specific IP.
	unboundIPv6 = "::"
)

// portDetails contains the details of a port.
type portDetails struct {
	InternalPort nat.Port
	HostPort     string
	Host         string
}

func (p *portDetails) Address() string {
	return net.JoinHostPort(p.Host, p.HostPort)
}

// hostPortMapping returns the exposed host port for the specified internalPort or the lowest port if its blank.
func hostPortMapping(ctx context.Context, internalPort nat.Port, pollInterval time.Duration, forceIPv4LocalHost bool, protocol string, target StrategyTarget) (*portDetails, error) {
	ipAddress, err := target.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("host: %w", err)
	}

	// Workaround for IPv6 docker bugs:
	// - https://github.com/moby/moby/issues/42442
	// - https://github.com/moby/moby/issues/42375
	if forceIPv4LocalHost {
		ipAddress = strings.Replace(ipAddress, "localhost", "127.0.0.1", 1)
	}

	var port *portDetails
	if internalPort == "" {
		port, err = lowestPort(ctx, pollInterval, protocol, target)
	} else {
		port, err = specifiedPort(ctx, internalPort, pollInterval, target)
	}

	if err != nil {
		return nil, err
	}

	// Ensure the target is still running.
	if err := checkTarget(ctx, target); err != nil {
		return nil, err
	}

	if port.Host == unboundIPv4 || port.Host == unboundIPv6 {
		port.Host = ipAddress
	}

	return port, nil
}

// lowestPort returns the lowest host port exposed by the container.
func lowestPort(ctx context.Context, pollInterval time.Duration, protocol string, target StrategyTarget) (*portDetails, error) {
	var lastExposedPorts int
	for {
		// Find the lowest numbered exposed tcp port.
		inspect, err := target.Inspect(ctx)
		if err != nil {
			return nil, fmt.Errorf("inspect container: %w", err)
		}

		if inspect.ContainerJSONBase.HostConfig.NetworkMode.IsHost() {
			return nil, fmt.Errorf("unable to determine port: network mode %q", inspect.ContainerJSONBase.HostConfig.NetworkMode)
		}

		var exposedPorts int
		var lowestPort nat.Port
		var binding nat.PortBinding
		for port, bindings := range inspect.NetworkSettings.Ports {
			if len(bindings) == 0 || protocol != "" && port.Proto() != protocol {
				// No bindings or not the protocol we are looking for.
				continue
			}

			exposedPorts++
			if lowestPort == "" || port.Int() < lowestPort.Int() {
				// Found a new lowest port.
				lowestPort = port
				binding = bindings[0]
			}
		}

		if lowestPort != "" && lastExposedPorts == exposedPorts {
			// Found the lowest port and no new ports were exposed.
			return &portDetails{
				InternalPort: lowestPort,
				HostPort:     binding.HostPort,
				Host:         binding.HostIP,
			}, nil
		}
		lastExposedPorts = exposedPorts

		select {
		case <-ctx.Done():
			// We use ErrPortNotFound for consistency with specifiedPort.
			return nil, fmt.Errorf("%w: %w", ctx.Err(), PortNotFoundErr(""))
		case <-time.After(pollInterval):
			if err := checkTarget(ctx, target); err != nil {
				return nil, err
			}
		}
	}
}

// specifiedPort returns the host port exposed by the container for the specified internal port.
func specifiedPort(ctx context.Context, internalPort nat.Port, pollInterval time.Duration, target StrategyTarget) (*portDetails, error) {
	for {
		inspect, err := target.Inspect(ctx)
		if err != nil {
			return nil, fmt.Errorf("inspect container: %w", err)
		}

		if inspect.ContainerJSONBase.HostConfig.NetworkMode.IsHost() {
			// The container is using the host network, so we assume
			// the port is already exposed.
			return &portDetails{
				InternalPort: internalPort,
				HostPort:     internalPort.Port(),
				Host:         unboundIPv4,
			}, nil
		}

		expectedPort := internalPort.Port()
		expectedProto := internalPort.Proto()
		for k, p := range inspect.NetworkSettings.Ports {
			if k.Port() != expectedPort || (expectedProto != "" && k.Proto() != expectedProto) {
				// Not the port we are looking for.
				continue
			}

			if len(p) == 0 {
				// The port is not exposed yet.
				continue
			}

			return &portDetails{
				InternalPort: k,
				HostPort:     p[0].HostPort,
				Host:         p[0].HostIP,
			}, nil
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %w", ctx.Err(), PortNotFoundErr(internalPort))
		case <-time.After(pollInterval):
			if err := checkTarget(ctx, target); err != nil {
				return nil, err
			}
		}
	}
}
