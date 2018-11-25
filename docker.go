package testcontainer

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainer-go/wait"
)

// RequestContainer is the input object used to get a running container.
type RequestContainer struct {
	Env          map[string]string
	ExportedPort []string
	Cmd          string
	RegistryCred string
	WaitingFor   wait.Strategy
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
// Deprecated: Use GetMappedPort
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
// Deprecated: Use GetContainerIpAddress
func (c *Container) GetIPAddress(ctx context.Context) (string, error) {
	inspect, err := c.inspectContainer(ctx)
	if err != nil {
		return "", err
	}
	return inspect.NetworkSettings.IPAddress, nil
}

// GetContainerIpAddress returns the ip address for the running container.
// Warning: this is based on your Docker host setting. Will fail if using an SSH tunnel
// You can use the "TC_HOST" env variable to set this yourself
func (c *Container) GetContainerIpAddress(ctx context.Context) (string, error) {
	host, err := daemonHost(*c.client)
	if err != nil {
		return "", err
	}
	return host, nil
}

// GetMappedPort returns the port reachable via the GetContainerIpAddress.
func (c *Container) GetMappedPort(ctx context.Context, port int) (int, error) {
	inspect, err := c.inspectContainer(ctx)
	if err != nil {
		return 0, err
	}

	for k, p := range inspect.NetworkSettings.Ports {
		if k.Port() == strconv.Itoa(port) {
			intp, err := strconv.Atoi(p[0].HostPort)
			if err != nil {
				return 0, err
			}
			return intp, nil
		}
	}
	return 0, nil
}

// GetHostEndpoint returns the IP address and the port exposed on the host machine.
// Deprecated: Use GetMappedPort
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

// daemonHost gets the host or ip of the Docker daemon where ports are exposed on
// Warning: this is based on your Docker host setting. Will fail if using an SSH tunnel
// You can use the "TC_HOST" env variable to set this yourself
func daemonHost(cli client.Client) (string, error) {
	host, exists := os.LookupEnv("TC_HOST")
	if exists {
		return host, nil
	}

	// infer from Docker host
	url, err := url.Parse(cli.DaemonHost())
	if err != nil {
		return "", err
	}

	switch url.Scheme {
	case "http", "https", "tcp":
		return url.Hostname(), nil
	case "unix", "npipe":
		// todo: get gateway address if inside container
		return "localhost", nil
	}

	return "", errors.New("Could not determine host through env or docker host")
}
