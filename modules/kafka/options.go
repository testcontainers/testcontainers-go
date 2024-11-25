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
func WithListener(listeners ...Listener) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if err := validateListeners(listeners...); err != nil {
			return fmt.Errorf("validate listeners: %w", err)
		}

		applyListenersToEnv(req, listeners...)

		return nil
	}
}

func applyListenersToEnv(req *testcontainers.GenericContainerRequest, listeners ...Listener) {
	if len(listeners) == 0 {
		return
	}

	req.Env["KAFKA_LISTENERS"] = "CONTROLLER://0.0.0.0:9094, PLAINTEXT://0.0.0.0:9093"
	req.Env["KAFKA_REST_BOOTSTRAP_SERVERS"] = "CONTROLLER://0.0.0.0:9094, PLAINTEXT://0.0.0.0:9093"
	req.Env["KAFKA_LISTENER_SECURITY_PROTOCOL_MAP"] = "CONTROLLER:PLAINTEXT, PLAINTEXT:PLAINTEXT"

	// expect first listener has common network between kafka instances
	req.Env["KAFKA_INTER_BROKER_LISTENER_NAME"] = listeners[0].Name

	// expect small number of listeners, so joins is okay
	for _, item := range listeners {
		req.Env["KAFKA_LISTENERS"] = strings.Join(
			[]string{
				req.Env["KAFKA_LISTENERS"],
				item.Name + "://0.0.0.0:" + item.Port,
			},
			",",
		)

		req.Env["KAFKA_REST_BOOTSTRAP_SERVERS"] = req.Env["KAFKA_LISTENERS"]

		req.Env["KAFKA_LISTENER_SECURITY_PROTOCOL_MAP"] = strings.Join(
			[]string{
				req.Env["KAFKA_LISTENER_SECURITY_PROTOCOL_MAP"],
				item.Name + ":" + "PLAINTEXT",
			},
			",",
		)
	}
}

func validateListeners(listeners ...Listener) error {
	// Trim
	for i := 0; i < len(listeners); i++ {
		listeners[i].Name = strings.ToUpper(strings.Trim(listeners[i].Name, " "))
		listeners[i].Host = strings.Trim(listeners[i].Host, " ")
		listeners[i].Port = strings.Trim(listeners[i].Port, " ")
	}

	// Validate
	ports := make(map[string]struct{}, len(listeners)+2)
	names := make(map[string]struct{}, len(listeners)+2)

	// check for default listeners
	ports["9094"] = struct{}{}
	ports["9093"] = struct{}{}

	// check for default listeners
	names["CONTROLLER"] = struct{}{}
	names["PLAINTEXT"] = struct{}{}

	for _, item := range listeners {
		if _, exists := names[item.Name]; exists {
			return fmt.Errorf("duplicate of listener name: %s", item.Name)
		}
		names[item.Name] = struct{}{}

		if _, exists := ports[item.Port]; exists {
			return fmt.Errorf("duplicate of listener port: %s", item.Port)
		}
		ports[item.Port] = struct{}{}
	}

	return nil
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
