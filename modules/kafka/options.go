package kafka

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	// Listeners is a list of custom listeners that can be provided to access the
	// containers form within docker networks
	Listeners []KafkaListener
}

func defaultOptions() options {
	return options{
		Listeners: make([]KafkaListener, 0),
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the Redpanda container.
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

// WithListener adds a custom listener to the Redpanda containers. Listener
// will be aliases to all networks, so they can be accessed from within docker
// networks. At leas one network must be attached to the container, if not an
// error will be thrown when starting the container.
func WithListener(listeners []KafkaListener) Option {
	return func(o *options) {
		o.Listeners = append(o.Listeners, listeners...)
	}
}

func externalListener(ctx context.Context, c testcontainers.Container) (KafkaListener, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return KafkaListener{}, err
	}

	port, err := c.MappedPort(ctx, publicPort)
	if err != nil {
		return KafkaListener{}, err
	}

	return KafkaListener{
		Name: "EXTERNAL",
		Ip:   host,
		Port: port.Port(),
	}, nil
}

func internalListener(ctx context.Context, c testcontainers.Container) (KafkaListener, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return KafkaListener{}, err
	}

	return KafkaListener{
		Name: "INTERNAL",
		Ip:   host,
		Port: "9092",
	}, nil
}
