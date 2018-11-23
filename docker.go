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
	ID string
	// Cache to retrieve container infromation without re-fetching them from dockerd
	raw    *types.ContainerJSON
	client *client.Client
}

// LivenessCheckPorts returns the exposed ports for the container.
func (c *Container) LivenessCheckPorts(ctx context.Context) (nat.PortSet, error) {
	inspect, err := c.inspectContainer(ctx)
	if err != nil {
		return nil, err
	}
	return inspect.Config.ExposedPorts, nil
}

// Terminate is used to kill the container. It is usally triggered by as defer function.
func (c *Container) Terminate(ctx context.Context, t *testing.T) error {
	return c.client.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
		Force: true,
	})
}

func (c *Container) inspectContainer(ctx context.Context) (*types.ContainerJSON, error) {
	if c.raw != nil {
		return c.raw, nil
	}
	inspect, err := c.client.ContainerInspect(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.raw = &inspect
	return c.raw, nil
}

// GetIPAddress returns the ip address for the running container.
func (c *Container) GetIPAddress(ctx context.Context) (string, error) {
	inspect, err := c.inspectContainer(ctx)
	if err != nil {
		return "", err
	}
	return inspect.NetworkSettings.IPAddress, nil
}

// GetHostEndpoint returns the IP address and the port exposed on the host machine.
func (c *Container) GetHostEndpoint(ctx context.Context, port string) (string, string, error) {
	inspect, err := c.inspectContainer(ctx)
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

// RunContainer takes a RequestContainer as input and it runs a container via the docker sdk
func RunContainer(ctx context.Context, containerImage string, input RequestContainer) (*Container, error) {
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

	_, _, err = cli.ImageInspectWithRaw(ctx, containerImage)
	if err != nil {
		if client.IsErrNotFound(err) {
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
		} else {
			return nil, err
		}
	}

	hostConfig := &container.HostConfig{
		PortBindings: exposedPortMap,
	}

	resp, err := cli.ContainerCreate(ctx, dockerInput, hostConfig, nil, "")
	if err != nil {
		return nil, err
	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}
	containerInstance := &Container{
		ID:     resp.ID,
		client: cli,
	}

	// if a WaitStrategy has been specified, wait before returning
	if input.WaitingFor != nil {
		if err := input.WaitingFor.WaitUntilReady(ctx, containerInstance); err != nil {
			// return containerInstance for termination
			return containerInstance, err
		}
	}
	return containerInstance, nil
}
