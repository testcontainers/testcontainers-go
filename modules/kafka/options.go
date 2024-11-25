package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	// Listeners is a list of custom listeners that can be provided to access the
	// containers form within docker networks
	Listeners []Listener
}

func defaultOptions() options {
	return options{
		Listeners: make([]Listener, 0),
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the Kafka container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

func WithClusterID(clusterID string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["CLUSTER_ID"] = clusterID
		return nil
	}
}

// WithListener adds a custom listener to the Kafka containers. Listener
// will be aliases to all networks, so they can be accessed from within docker
// networks. At least one network must be attached to the container, if not an
// error will be thrown when starting the container.
// This options sanitizes the listener names and ports, so they are in the
// correct format: name is uppercase and trimmed, and port is trimmed.
func WithListener(listeners []Listener) Option {
	// Trim
	for i := 0; i < len(listeners); i++ {
		listeners[i].Name = strings.ToUpper(strings.Trim(listeners[i].Name, " "))
		listeners[i].Host = strings.Trim(listeners[i].Host, " ")
		listeners[i].Port = strings.Trim(listeners[i].Port, " ")
	}

	return func(o *options) {
		o.Listeners = append(o.Listeners, listeners...)
	}
}

func plainTextListener(ctx context.Context, c testcontainers.Container) (Listener, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return Listener{}, err
	}

	port, err := c.MappedPort(ctx, publicPort)
	if err != nil {
		return Listener{}, err
	}

	return Listener{
		Name: "PLAINTEXT",
		Host: host,
		Port: port.Port(),
	}, nil
}

func brokerListener(ctx context.Context, c testcontainers.Container) (Listener, error) {
	inspect, err := c.Inspect(ctx)
	if err != nil {
		return Listener{}, fmt.Errorf("inspect: %w", err)
	}

	return Listener{
		Name: "BROKER",
		Host: inspect.Config.Hostname,
		Port: "9092",
	}, nil
}
