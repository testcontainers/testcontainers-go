package nats

import (
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

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*CmdOption)(nil)

// CmdOption is an option for the NATS container.
type CmdOption func(opts *options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o CmdOption) Customize(req *testcontainers.GenericContainerRequest) {
	// NOOP to satisfy interface.
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
