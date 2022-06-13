package testcontainers

import (
	"context"
)

type ContainerisedDockerComposeOptions struct {
	Context  context.Context
	Provider ContainerProvider
}

type ContainerisedDockerCompose struct {
	ContainerisedDockerComposeOptions
	Env     map[string]string
	Context context.Context
}

func (c *ContainerisedDockerCompose) Invoke() ExecError {
	req := ContainerRequest{
		Image:      "docker/compose:1.29.2",
		Env:        c.Env,
		Mounts:     Mounts(BindMount("/var/run/docker.sock", "/var/run/docker.sock")),
		SkipReaper: true,
		AutoRemove: true,
	}

	_, err := c.Provider.RunContainer(c.Context, req)
	if err != nil {
		return ExecError{}
	}

	// TODO: waiting until finished

	return ExecError{}
}

func NewContainerisedDockerCompose(filePaths []string, identifier string, options ContainerisedDockerComposeOptions) *ContainerisedDockerCompose {
	composeFile := ""

	return &ContainerisedDockerCompose{
		ContainerisedDockerComposeOptions: options,
		Env: map[string]string{
			"ENV_PROJECT_NAME": identifier,
			"ENV_COMPOSE_FILE": composeFile,
		},
	}
}
