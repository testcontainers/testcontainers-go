package nats

import (
	"io"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	CmdArgs map[string]string
}

func defaultOptions() options {
	return options{
		CmdArgs: make(map[string]string, 0),
	}
}

// Compiler check to ensure that CmdOption implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*CmdOption)(nil)

// CmdOption is an option for the NATS container.
type CmdOption func(opts *options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o CmdOption) Customize(_ *testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

func WithUsername(username string) CmdOption {
	return func(o *options) {
		o.CmdArgs["user"] = username
	}
}

func WithPassword(password string) CmdOption {
	return func(o *options) {
		o.CmdArgs["pass"] = password
	}
}

// WithArgument adds an argument and its value to the NATS container.
// The argument flag does not need to include the dashes.
func WithArgument(flag string, value string) CmdOption {
	flag = strings.ReplaceAll(flag, "--", "") // remove all dashes to make it easier to use

	return func(o *options) {
		o.CmdArgs[flag] = value
	}
}

// WithConfigFile pass a content of io.Reader to the NATS container as /etc/nats.conf
// Changing the connectivity (listen address or ports) can break the container setup.
func WithConfigFile(config io.Reader) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if config != nil {
			req.Cmd = append(req.Cmd, "-config", "/etc/nats.conf")
			req.Files = append(
				req.Files,
				testcontainers.ContainerFile{
					Reader:            config,
					ContainerFilePath: "/etc/nats.conf",
					FileMode:          0o644,
				},
			)
		}
		return nil
	}
}
