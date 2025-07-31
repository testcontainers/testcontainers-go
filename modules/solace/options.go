package solace

import (
	"errors"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	vpn      string
	username string
	password string
	services []Service           // enabled services
	queues   map[string][]string // queueName -> topics
	shmSize  int64
}

func defaultOptions() options {
	return options{
		vpn:      "default",
		username: "root",
		password: "password",
		services: defaultServices,
		shmSize:  1 << 30, // 1 GiB
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

// WithServices configures the services to be exposed with their wait strategies
func WithServices(srv ...Service) Option {
	return func(o *options) error {
		if len(srv) == 0 {
			return errors.New("at least one service must be specified")
		}
		// Clear existing services and use only the specified ones
		o.services = srv
		return nil
	}
}

// WithCredentials sets the client credentials (username, password)
func WithCredentials(username, password string) Option {
	return func(o *options) error {
		o.username = username
		o.password = password
		return nil
	}
}

// WithVPN sets the VPN name
func WithVPN(vpn string) Option {
	return func(o *options) error {
		o.vpn = vpn
		return nil
	}
}

// WithQueue subscribes a given topic to a queue (for SMF/AMQP testing)
func WithQueue(queueName, topic string) Option {
	return func(o *options) error {
		if o.queues == nil {
			o.queues = make(map[string][]string)
		}
		o.queues[queueName] = append(o.queues[queueName], topic)
		return nil
	}
}

// WithShmSize sets the size of the /dev/shm volume
func WithShmSize(size int64) Option {
	return func(o *options) error {
		o.shmSize = size
		return nil
	}
}
