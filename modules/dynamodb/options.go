package dynamodb

import "github.com/testcontainers/testcontainers-go"

type options struct {
	cmd []string
}

func defaultOptions() options {
	return options{}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the DynamoDB container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithSharedDB allows container reuse between successive runs. Data will be persisted
func WithSharedDB() Option {
	return func(o *options) error {
		o.cmd = append(o.cmd, "-sharedDb")

		return nil
	}
}

// WithDisableTelemetry - DynamoDB local will not send any telemetry
func WithDisableTelemetry() Option {
	return func(o *options) error {
		// if other flags (e.g. -sharedDb) exist, append to them
		o.cmd = append(o.cmd, "-disableTelemetry")

		return nil
	}
}
