package testcontainers

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type ContainerDependency struct {
	Request      ContainerRequest
	EnvKey       string
	CallbackFunc func(Container)
}

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
	})
	if err != nil {
		return nil, err
	}

	c.CallbackFunc(dependency)
	return dependency, nil
}

func resolveDNSName(ctx context.Context, container Container, network string) (string, error) {
	networkAlias, err := container.NetworkAliases(ctx)
	if err != nil {
		return "", err
	}
	aliases := networkAlias[network]
	if len(aliases) == 0 {
		return "", errors.New("could not retrieve network alias for dependency container")
	}
	return aliases[0], nil
}

func cleanupDependency(ctx context.Context, dependencies []Container, network *DockerNetwork) error {
	if network == nil {
		return nil
	}

	for _, dependency := range dependencies {
		err := dependency.Terminate(ctx)
		if err != nil {
			return err
		}
	}
	return network.Remove(ctx)
}

var defaultDependencyHook = func(dockerInput *container.Config, client client.APIClient) ContainerLifecycleHooks {
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
						cleanupErr := cleanupDependency(ctx, depContainers, depNetwork)
						Logger.Printf("Could not cleanup dependencies after an error occured: %v", cleanupErr)
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
					name, err := resolveDNSName(ctx, container, depNetwork.Name)
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
				if depNetwork != nil {
					return client.NetworkConnect(ctx, depNetwork.ID, container.GetContainerID(), nil)
				}
				return nil
			},
		},
		PostTerminates: []ContainerHook{
			func(ctx context.Context, container Container) error {
				return cleanupDependency(ctx, depContainers, depNetwork)
			},
		},
	}
}
