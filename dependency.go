package testcontainers

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

// ContainerDependency represents a reliance that a container has on another container.
type ContainerDependency struct {
	Request ContainerRequest
	EnvKey  string
	// CallbackFunc is called after the dependency container is started.
	CallbackFunc func(Container)
}

// NewContainerDependency can be used to define a dependency and the environment variable that
// will be used to pass the DNS name to the parent container.
func NewContainerDependency(containerRequest ContainerRequest, envKey string) *ContainerDependency {
	return &ContainerDependency{
		Request:      containerRequest,
		EnvKey:       envKey,
		CallbackFunc: func(c Container) {},
	}
}

func (c *ContainerDependency) WithCallback(callbackFunc func(Container)) *ContainerDependency {
	c.CallbackFunc = callbackFunc
	return c
}

func (c *ContainerDependency) StartDependency(ctx context.Context, network string) (Container, error) {
	c.Request.Networks = append(c.Request.Networks, network)
	dependency, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: c.Request,
		Started:          true,
		Reuse:            c.Request.Name != "", // reuse a running dependency container if a name is provided.
	})
	if err != nil {
		return nil, err
	}

	c.CallbackFunc(dependency)
	return dependency, nil
}

func resolveDNSName(ctx context.Context, container Container, network *DockerNetwork) (string, error) {
	curNetworks, err := container.Networks(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: could not retrieve networks for dependency container", err)
	}
	// The container may not be connected to the network if it was reused.
	if slices.Index(curNetworks, network.Name) == -1 {
		err = network.provider.client.NetworkConnect(ctx, network.ID, container.GetContainerID(), nil)
		if err != nil {
			return "", fmt.Errorf("%w: could not connect dependency container to network", err)
		}
	}

	networkAlias, err := container.NetworkAliases(ctx)
	if err != nil {
		return "", err
	}

	aliases := networkAlias[network.Name]
	if len(aliases) == 0 {
		return "", errors.New("could not retrieve network alias for dependency container")
	}
	return aliases[0], nil
}

func cleanupDependencyNetwork(ctx context.Context, dependencies []Container, network *DockerNetwork) error {
	if network == nil {
		return nil
	}

	for _, dependency := range dependencies {
		err := network.provider.client.NetworkDisconnect(ctx, network.ID, dependency.GetContainerID(), true)
		if err != nil {
			return err
		}
	}
	defer network.provider.Close()
	return network.Remove(ctx)
}

var defaultDependencyHook = func(dockerInput *container.Config) ContainerLifecycleHooks {
	var depNetwork *DockerNetwork
	depContainers := make([]Container, 0)
	return ContainerLifecycleHooks{
		PreCreates: []ContainerRequestHook{
			func(ctx context.Context, req ContainerRequest) (err error) {
				if len(req.DependsOn) == 0 {
					return nil
				}
				defer func() {
					if err != nil {
						// clean up dependencies that were created if an error occurred.
						cleanupErr := cleanupDependencyNetwork(ctx, depContainers, depNetwork)
						if cleanupErr != nil {
							Logger.Printf("Could not cleanup dependencies after an error occured: %v", cleanupErr)
						}
					}
				}()

				net, err := GenericNetwork(ctx, GenericNetworkRequest{
					NetworkRequest: NetworkRequest{
						Driver:   Bridge,
						Labels:   GenericLabels(),
						Name:     fmt.Sprintf("testcontainer-dependency-%v", uuid.NewString()),
						Internal: false,
					},
				})
				depNetwork = net.(*DockerNetwork)
				if err != nil {
					return err
				}

				for _, dep := range req.DependsOn {
					if dep.EnvKey == "" {
						return errors.New("cannot create dependency with empty environment key.")
					}
					container, err := dep.StartDependency(ctx, depNetwork.Name)
					if err != nil {
						return err
					}
					depContainers = append(depContainers, container)
					name, err := resolveDNSName(ctx, container, depNetwork)
					if err != nil {
						return err
					}
					dockerInput.Env = append(dockerInput.Env, dep.EnvKey+"="+name)
				}
				return nil
			},
		},
		PostCreates: []ContainerHook{
			func(ctx context.Context, container Container) error {
				if depNetwork == nil {
					return nil
				}
				err := depNetwork.provider.client.NetworkConnect(ctx, depNetwork.ID, container.GetContainerID(), nil)
				defer depNetwork.provider.Close()
				return err
			},
		},
		PostTerminates: []ContainerHook{
			func(ctx context.Context, container Container) error {
				return cleanupDependencyNetwork(ctx, depContainers, depNetwork)
			},
		},
	}
}
