package testcontainers

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Dependency interface {
	StartDependency(context.Context, string) (string, error)
}

type ContainerDependency struct {
	request    ContainerRequest
	envKey     string
	callbackFn func(Container)
}

func (c *ContainerDependency) StartDependency(ctx context.Context, network string) (string, error) {
	c.request.Networks = append(c.request.Networks, network)
	dependency, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: c.request,
		Started:          true,
	})
	if err != nil {
		return "", err
	}

	c.callbackFn(dependency)
	networkAlias, err := dependency.NetworkAliases(ctx)
	if err != nil {
		return "", err
	}
	aliases := networkAlias[network]
	if len(aliases) == 0 {
		return "", errors.New("could not retrieve network alias for dependency container")
	}
	return aliases[0], nil
}

func NewContainerDependency(containerRequest ContainerRequest, envVar string) *ContainerDependency {
	return &ContainerDependency{
		request:    containerRequest,
		envKey:     envVar,
		callbackFn: func(c Container) {},
	}
}

var defaultDependencyHook = func(dockerInput *container.Config, client client.APIClient) ContainerLifecycleHooks {
	var depNetwork *DockerNetwork
	return ContainerLifecycleHooks{
		PreCreates: []ContainerRequestHook{
			func(ctx context.Context, req ContainerRequest) error {
				if len(req.DependsOn) == 0 {
					return nil
				}

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
					name, err := dep.StartDependency(ctx, depNetwork.Name)
					if err != nil {
						return err
					}
					envKey := dep.(*ContainerDependency).envKey
					dockerInput.Env = append(dockerInput.Env, envKey+"="+name)
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
				return depNetwork.Remove(ctx)
			},
		},
	}
}
