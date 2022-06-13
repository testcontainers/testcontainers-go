package testcontainers

import (
	"context"
)

type ContainerisedDockerCompose struct {
	Provider ContainerProvider
	Env      map[string]string
	Context  context.Context
}

func (c *ContainerisedDockerCompose) Invoke() ExecError {
	req := ContainerRequest{
		Image:      reaperImage("docker/compose:1.29.2"),
		Env:        c.Env,
		Mounts:     Mounts(BindMount("/var/run/docker.sock", "/var/run/docker.sock")),
		SkipReaper: true,
		AutoRemove: true,
	}

	c.Provider.CreateContainer(c.Context, req)
	return ExecError{}
}

func NewContainerisedDockerCompose(filePaths []string, identifier string, ctx context.Context) *ContainerisedDockerCompose {
	composeFile := ""

	return &ContainerisedDockerCompose{
		Env: map[string]string{
			"ENV_PROJECT_NAME": identifier,
			"ENV_COMPOSE_FILE": composeFile,
		},
		Context: ctx,
	}
}
