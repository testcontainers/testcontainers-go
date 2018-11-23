package testcontainer

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainer-go/wait"
)

// RequestContainer is the input object used to get a running container.
type RequestContainer struct {
	Env          map[string]string
	ExportedPort []string
	Cmd          string
	RegistryCred string
	WaitingFor   wait.WaitStrategy
}

// Container is the struct used to represent a single container.
type Container struct {
	// Container ID from Docker
	ID      string
	started bool
	// Cache to retrieve container infromation without re-fetching them from dockerd
	raw *types.ContainerJSON
}

// LivenessCheckPorts returns the exposed ports for the container.
func (c *Container) LivenessCheckPorts(ctx context.Context) (nat.PortSet, error) {
	inspect, err := inspectContainer(ctx, c)
	if err != nil {
		return nil, err
	}
	return inspect.Config.ExposedPorts, nil
}

// Terminate is used to kill the container. It is usally triggered by as defer function.
func (c *Container) Terminate(ctx context.Context, t *testing.T) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Error(err)
		return err
	}
	return cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
		Force: true,
	})
}

func inspectContainer(ctx context.Context, c *Container) (*types.ContainerJSON, error) {
	if c.raw != nil {
		return c.raw, nil
	}
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	inspect, err := cli.ContainerInspect(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.raw = &inspect
	return c.raw, nil
}

// GetIPAddress returns the ip address for the running container.
func (c *Container) GetIPAddress(ctx context.Context) (string, error) {
	inspect, err := inspectContainer(ctx, c)
	if err != nil {
		return "", err
	}
	return inspect.NetworkSettings.IPAddress, nil
}

// GetHostEndpoint returns the IP address and the port exposed on the host machine.
func (c *Container) GetHostEndpoint(ctx context.Context, port string) (string, string, error) {
	inspect, err := inspectContainer(ctx, c)
	if err != nil {
		return "", "", err
	}

	portSet, _, err := nat.ParsePortSpecs([]string{port})
	if err != nil {
		return "", "", err
	}

	for p := range portSet {
		ports, ok := inspect.NetworkSettings.Ports[p]
		if !ok {
			return "", "", fmt.Errorf("port %s not found", port)
		}
		if len(ports) == 0 {
			return "", "", fmt.Errorf("port %s not found", port)
		}

		return ports[0].HostIP, ports[0].HostPort, nil

	}

	return "", "", fmt.Errorf("port %s not found", port)
}

// Run is the function that starts the container. It can executed once
func (c *Container) Run(ctx context.Context, waitFor wait.WaitStrategy) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	if c.started == true {
		return fmt.Errorf("you can start a container only one.")
	}
	if err := cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	c.started = true

	if err := waitFor.WaitUntilReady(ctx, c); err != nil {
		return err
	}
	return nil
}

// RunAndWait is the function that starts the container and it uses a waiting strategy. It can executed once
func (c *Container) RunAndWait(ctx context.Context, waitFor wait.WaitStrategy) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	if c.started == true {
		return fmt.Errorf("you can start a container only one.")
	}
	if err := cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	c.started = true

	if err := waitFor.WaitUntilReady(ctx, c); err != nil {
		return err
	}
	return nil
}

// CreateContainer takes a RequestContainer as input and it creates a container via the docker sdk
func CreateContainer(ctx context.Context, containerImage string, input RequestContainer) (*Container, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	exposedPortSet, exposedPortMap, err := nat.ParsePortSpecs(input.ExportedPort)
	if err != nil {
		return nil, err
	}

	env := []string{}
	for envKey, envVar := range input.Env {
		env = append(env, envKey+"="+envVar)
	}

	dockerInput := &container.Config{
		Image:        containerImage,
		Env:          env,
		ExposedPorts: exposedPortSet,
	}

	if input.Cmd != "" {
		dockerInput.Cmd = strings.Split(input.Cmd, " ")
	}

	pullOpt := types.ImagePullOptions{}
	if input.RegistryCred != "" {
		pullOpt.RegistryAuth = input.RegistryCred
	}
	pull, err := cli.ImagePull(ctx, dockerInput.Image, pullOpt)
	if err != nil {
		return nil, err
	}
	defer pull.Close()

	// download of docker image finishes at EOF of the pull request
	_, err = ioutil.ReadAll(pull)
	if err != nil {
		return nil, err
	}

	hostConfig := &container.HostConfig{
		PortBindings: exposedPortMap,
	}

	resp, err := cli.ContainerCreate(ctx, dockerInput, hostConfig, nil, "")
	if err != nil {
		return nil, err
	}
	containerInstance := &Container{
		ID:      resp.ID,
		started: false,
	}
	return containerInstance, nil
}

// RunContainer takes a RequestContainer as input and it creates and runs a container via the docker sdk
func RunContainer(ctx context.Context, containerImage string, input RequestContainer) (*Container, error) {
	containerInstance, err := CreateContainer(ctx, containerImage, input)
	if err != nil {
		return nil, err
	}
	if input.WaitingFor != nil {
		err = containerInstance.RunAndWait(ctx, input.WaitingFor)
		if err != nil {
			return nil, err
		}
	}
	return containerInstance, nil
}
