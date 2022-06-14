package testcontainers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ContainerisedDockerComposeOptions struct {
	Context  context.Context
	Provider ContainerProvider
}

type ContainerisedDockerCompose struct {
	ContainerisedDockerComposeOptions
	Env     map[string]string
	Context context.Context
	// Pwd is a path to the directory where compose files located.
	Pwd string
	// ContainerPwd is a path to the directory in the container where Pwd mounted.
	ContainerPwd string
}

func (c *ContainerisedDockerCompose) Invoke() ExecError {
	req := ContainerRequest{
		Image: "docker/compose:1.29.2",
		Env:   c.Env,
		Mounts: Mounts(
			BindMount(
				coalesce(os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE"), "/var/run/docker.sock"),
				"/var/run/docker.sock",
			),
			BindMount(c.Pwd, ContainerMountTarget(c.ContainerPwd)),
		),
		Cmd: []string{"docker-compose"},
	}

	container, err := c.Provider.RunContainer(c.Context, req)
	if err != nil {
		return ExecError{
			Command: req.Cmd,
			Error:   err,
		}
	}

	state, err := container.State(c.Context)

	for err == nil && state.FinishedAt == "" {
		time.Sleep(60000)
		state, err = container.State(c.Context)
	}

	if state.Status == "died" {
		err = fmt.Errorf(state.Error)
	}

	if state.Status == "exited" && state.ExitCode != 0 {
		err = fmt.Errorf(state.Error)
	}

	return ExecError{
		Command: req.Cmd,
		Error:   err,
	}
}

func NewContainerisedDockerCompose(filePaths []string, identifier string, options ContainerisedDockerComposeOptions) *ContainerisedDockerCompose {
	containerPwd := "/app"
	pwd, err := filepath.Abs(filepath.Dir(filePaths[0]))

	if err != nil {
		return nil
	}

	var composeFiles []string

	for _, p := range filePaths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil
		}

		if !strings.HasPrefix(absPath, pwd+string(filepath.Separator)) {
			return nil
		}

		composeFiles = append(composeFiles, containerPwd+"/"+filepath.ToSlash(absPath[len(pwd)+1:]))
	}

	return &ContainerisedDockerCompose{
		Pwd:                               pwd,
		ContainerPwd:                      containerPwd,
		ContainerisedDockerComposeOptions: options,
		Env: map[string]string{
			"COMPOSE_PROJECT_NAME": identifier,
			"COMPOSE_FILE":         strings.Join(composeFiles, ":"),
		},
	}
}
