package toxiproxy

import (
	"errors"
	"io"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	portRange int
}

func defaultOptions() options {
	return options{
		portRange: defaultPortRange,
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

// WithPortRange sets the port range for the Toxiproxy container.
// Default port range is 31.
func WithPortRange(portRange int) Option {
	return func(o *options) error {
		if portRange < 1 {
			return errors.New("port range must be greater than 0")
		}

		o.portRange = portRange
		return nil
	}
}

// WithConfigFile sets the config file for the Toxiproxy container, copying
// the file to the "/tmp/tc-toxiproxy.json" path. It also appends the "-host=0.0.0.0"
// and "-config=/tmp/tc-toxiproxy.json" flags to the command line.
// The config file is a JSON file that contains the configuration for the Toxiproxy container,
// and it is not validated by the Toxiproxy container.
func WithConfigFile(r io.Reader) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Files = append(req.Files, testcontainers.ContainerFile{
			Reader:            r,
			ContainerFilePath: "/tmp/tc-toxiproxy.json",
			FileMode:          0o644,
		})

		req.Cmd = append(req.Cmd, "-host=0.0.0.0", "-config=/tmp/tc-toxiproxy.json")
		return nil
	}
}
