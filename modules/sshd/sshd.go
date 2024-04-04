package sshd

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// defaultImage is the default image used for the SSHD container
const defaultImage = "testcontainers/sshd:1.1.0"

// defaultPort is the default SSHD port
const defaultPort = "22/tcp"

// defaultCommand is the default SSHD command
var defaultCommand = []string{
	"sh",
	"-c",
	"echo \"root:$ROOT_PASSWORD\" | chpasswd && /usr/sbin/sshd -D",
}

// SSHDContainer represents the SSHD container
type SSHDContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the SSHD container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*SSHDContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        defaultImage,
		ExposedPorts: []string{defaultPort},
		// WaitingFor:   wait.ForLog("* Ready to accept connections"),
		Cmd: defaultCommand,
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	// Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&settings)
		}
		opt.Customize(&genericContainerReq)
	}

	applyRootPassword(settings.RootPassword)(&genericContainerReq)
	applyCommandOptions(settings.Options)(&genericContainerReq)

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &SSHDContainer{Container: container}, nil
}

func applyRootPassword(rootPassword string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["ROOT_PASSWORD"] = rootPassword
	}
}

func applyCommandOptions(commandOptions []string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		for _, commandOption := range commandOptions {
			commandOption = "-o " + commandOption
			req.Cmd = append(req.Cmd, commandOption)
		}
	}
}
