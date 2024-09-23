package testcontainers

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPreCreateModifierHook(t *testing.T) {
	ctx := context.Background()

	provider, err := NewDockerProvider()
	require.NoError(t, err)
	defer provider.Close()

	t.Run("No exposed ports", func(t *testing.T) {
		// reqWithModifiers {
		req := ContainerRequest{
			Image: nginxAlpineImage, // alpine image does expose port 80
			ConfigModifier: func(config *container.Config) {
				config.Env = []string{"a=b"}
			},
			Mounts: ContainerMounts{
				{
					Source: DockerVolumeMountSource{
						Name: "appdata",
						VolumeOptions: &mount.VolumeOptions{
							Labels: GenericLabels(),
						},
					},
					Target: "/data",
				},
			},
			HostConfigModifier: func(hostConfig *container.HostConfig) {
				hostConfig.PortBindings = nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1",
							HostPort: "2",
						},
					},
				}
			},
			EnpointSettingsModifier: func(endpointSettings map[string]*network.EndpointSettings) {
				endpointSettings["a"] = &network.EndpointSettings{
					Aliases: []string{"b"},
					Links:   []string{"link1", "link2"},
				}
			},
		}
		// }

		// define empty inputs to be overwritten by the pre create hook
		inputConfig := &container.Config{
			Image: req.Image,
		}
		inputHostConfig := &container.HostConfig{}
		inputNetworkingConfig := &network.NetworkingConfig{}

		err = provider.preCreateContainerHook(ctx, req, inputConfig, inputHostConfig, inputNetworkingConfig)
		require.NoError(t, err)

		// assertions

		assert.Equal(
			t,
			[]string{"a=b"},
			inputConfig.Env,
			"Docker config's env should be overwritten by the modifier",
		)
		assert.Equal(t,
			nat.PortSet(nat.PortSet{"80/tcp": struct{}{}}),
			inputConfig.ExposedPorts,
			"Docker config's exposed ports should be overwritten by the modifier",
		)
		assert.Equal(
			t,
			[]mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: "appdata",
					Target: "/data",
					VolumeOptions: &mount.VolumeOptions{
						Labels: GenericLabels(),
					},
				},
			},
			inputHostConfig.Mounts,
			"Host config's mounts should be mapped to Docker types",
		)

		assert.Equal(t, nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostIP:   "",
					HostPort: "",
				},
			},
		}, inputHostConfig.PortBindings,
			"Host config's port bindings should be overwritten by the modifier",
		)

		assert.Equal(
			t,
			[]string{"b"},
			inputNetworkingConfig.EndpointsConfig["a"].Aliases,
			"Networking config's aliases should be overwritten by the modifier",
		)
		assert.Equal(
			t,
			[]string{"link1", "link2"},
			inputNetworkingConfig.EndpointsConfig["a"].Links,
			"Networking config's links should be overwritten by the modifier",
		)
	})

	t.Run("No exposed ports and network mode IsContainer", func(t *testing.T) {
		req := ContainerRequest{
			Image: nginxAlpineImage, // alpine image does expose port 80
			HostConfigModifier: func(hostConfig *container.HostConfig) {
				hostConfig.PortBindings = nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1",
							HostPort: "2",
						},
					},
				}
				hostConfig.NetworkMode = "container:foo"
			},
		}

		// define empty inputs to be overwritten by the pre create hook
		inputConfig := &container.Config{
			Image: req.Image,
		}
		inputHostConfig := &container.HostConfig{}
		inputNetworkingConfig := &network.NetworkingConfig{}

		err = provider.preCreateContainerHook(ctx, req, inputConfig, inputHostConfig, inputNetworkingConfig)
		require.NoError(t, err)

		// assertions

		assert.Equal(
			t,
			nat.PortSet(nat.PortSet{}),
			inputConfig.ExposedPorts,
			"Docker config's exposed ports should be empty",
		)
		assert.Equal(t,
			nat.PortMap{},
			inputHostConfig.PortBindings,
			"Host config's portBinding should be empty",
		)
	})

	t.Run("Nil hostConfigModifier should apply default host config modifier", func(t *testing.T) {
		req := ContainerRequest{
			Image:       nginxAlpineImage, // alpine image does expose port 80
			AutoRemove:  true,
			CapAdd:      []string{"addFoo", "addBar"},
			CapDrop:     []string{"dropFoo", "dropBar"},
			Binds:       []string{"bindFoo", "bindBar"},
			ExtraHosts:  []string{"hostFoo", "hostBar"},
			NetworkMode: "networkModeFoo",
			Resources: container.Resources{
				Memory:   2048,
				NanoCPUs: 8,
			},
			HostConfigModifier: nil,
		}

		// define empty inputs to be overwritten by the pre create hook
		inputConfig := &container.Config{
			Image: req.Image,
		}
		inputHostConfig := &container.HostConfig{}
		inputNetworkingConfig := &network.NetworkingConfig{}

		err = provider.preCreateContainerHook(ctx, req, inputConfig, inputHostConfig, inputNetworkingConfig)
		require.NoError(t, err)

		// assertions

		assert.Equal(t, req.AutoRemove, inputHostConfig.AutoRemove, "Deprecated AutoRemove should come from the container request")
		assert.Equal(t, strslice.StrSlice(req.CapAdd), inputHostConfig.CapAdd, "Deprecated CapAdd should come from the container request")
		assert.Equal(t, strslice.StrSlice(req.CapDrop), inputHostConfig.CapDrop, "Deprecated CapDrop should come from the container request")
		assert.Equal(t, req.Binds, inputHostConfig.Binds, "Deprecated Binds should come from the container request")
		assert.Equal(t, req.ExtraHosts, inputHostConfig.ExtraHosts, "Deprecated ExtraHosts should come from the container request")
		assert.Equal(t, req.Resources, inputHostConfig.Resources, "Deprecated Resources should come from the container request")
	})

	t.Run("Request contains more than one network including aliases", func(t *testing.T) {
		networkName := "foo"
		net, err := provider.CreateNetwork(ctx, NetworkRequest{
			Name: networkName,
		})
		require.NoError(t, err)
		CleanupNetwork(t, net)

		dockerNetwork, err := provider.GetNetwork(ctx, NetworkRequest{
			Name: networkName,
		})
		require.NoError(t, err)

		req := ContainerRequest{
			Image:    nginxAlpineImage, // alpine image does expose port 80
			Networks: []string{networkName, "bar"},
			NetworkAliases: map[string][]string{
				"foo": {"foo1"}, // network aliases are needed at the moment there is a network
			},
		}

		// define empty inputs to be overwritten by the pre create hook
		inputConfig := &container.Config{
			Image: req.Image,
		}
		inputHostConfig := &container.HostConfig{}
		inputNetworkingConfig := &network.NetworkingConfig{}

		err = provider.preCreateContainerHook(ctx, req, inputConfig, inputHostConfig, inputNetworkingConfig)
		require.NoError(t, err)

		// assertions

		assert.Equal(
			t,
			req.NetworkAliases[networkName],
			inputNetworkingConfig.EndpointsConfig[networkName].Aliases,
			"Networking config's aliases should come from the container request",
		)
		assert.Equal(
			t,
			dockerNetwork.ID,
			inputNetworkingConfig.EndpointsConfig[networkName].NetworkID,
			"Networking config's network ID should be retrieved from Docker",
		)
	})

	t.Run("Request contains more than one network without aliases", func(t *testing.T) {
		networkName := "foo"
		net, err := provider.CreateNetwork(ctx, NetworkRequest{
			Name: networkName,
		})
		require.NoError(t, err)
		CleanupNetwork(t, net)

		dockerNetwork, err := provider.GetNetwork(ctx, NetworkRequest{
			Name: networkName,
		})
		require.NoError(t, err)

		req := ContainerRequest{
			Image:    nginxAlpineImage, // alpine image does expose port 80
			Networks: []string{networkName, "bar"},
		}

		// define empty inputs to be overwritten by the pre create hook
		inputConfig := &container.Config{
			Image: req.Image,
		}
		inputHostConfig := &container.HostConfig{}
		inputNetworkingConfig := &network.NetworkingConfig{}

		err = provider.preCreateContainerHook(ctx, req, inputConfig, inputHostConfig, inputNetworkingConfig)
		require.NoError(t, err)

		// assertions

		assert.Empty(
			t,
			inputNetworkingConfig.EndpointsConfig[networkName].Aliases,
			"Networking config's aliases should be empty",
		)
		assert.Equal(
			t,
			dockerNetwork.ID,
			inputNetworkingConfig.EndpointsConfig[networkName].NetworkID,
			"Networking config's network ID should be retrieved from Docker",
		)
	})

	t.Run("Request contains exposed port modifiers without protocol", func(t *testing.T) {
		req := ContainerRequest{
			Image: nginxAlpineImage, // alpine image does expose port 80
			HostConfigModifier: func(hostConfig *container.HostConfig) {
				hostConfig.PortBindings = nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "localhost",
							HostPort: "8080",
						},
					},
				}
			},
			ExposedPorts: []string{"80"},
		}

		// define empty inputs to be overwritten by the pre create hook
		inputConfig := &container.Config{
			Image: req.Image,
		}
		inputHostConfig := &container.HostConfig{}
		inputNetworkingConfig := &network.NetworkingConfig{}

		err = provider.preCreateContainerHook(ctx, req, inputConfig, inputHostConfig, inputNetworkingConfig)
		require.NoError(t, err)

		// assertions
		assert.Equal(t, "localhost", inputHostConfig.PortBindings["80/tcp"][0].HostIP)
		assert.Equal(t, "8080", inputHostConfig.PortBindings["80/tcp"][0].HostPort)
	})

	t.Run("Request contains exposed port modifiers with protocol", func(t *testing.T) {
		req := ContainerRequest{
			Image: nginxAlpineImage, // alpine image does expose port 80
			HostConfigModifier: func(hostConfig *container.HostConfig) {
				hostConfig.PortBindings = nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "localhost",
							HostPort: "8080",
						},
					},
				}
			},
			ExposedPorts: []string{"80/tcp"},
		}

		// define empty inputs to be overwritten by the pre create hook
		inputConfig := &container.Config{
			Image: req.Image,
		}
		inputHostConfig := &container.HostConfig{}
		inputNetworkingConfig := &network.NetworkingConfig{}

		err = provider.preCreateContainerHook(ctx, req, inputConfig, inputHostConfig, inputNetworkingConfig)
		require.NoError(t, err)

		// assertions
		assert.Equal(t, "localhost", inputHostConfig.PortBindings["80/tcp"][0].HostIP)
		assert.Equal(t, "8080", inputHostConfig.PortBindings["80/tcp"][0].HostPort)
	})
}

func TestMergePortBindings(t *testing.T) {
	type arg struct {
		configPortMap nat.PortMap
		parsedPortMap nat.PortMap
		exposedPorts  []string
	}
	cases := []struct {
		name     string
		arg      arg
		expected nat.PortMap
	}{
		{
			name: "empty ports",
			arg: arg{
				configPortMap: nil,
				parsedPortMap: nil,
				exposedPorts:  nil,
			},
			expected: map[nat.Port][]nat.PortBinding{},
		},
		{
			name: "config port map but not exposed",
			arg: arg{
				configPortMap: map[nat.Port][]nat.PortBinding{
					"80/tcp": {{HostIP: "1", HostPort: "2"}},
				},
				parsedPortMap: nil,
				exposedPorts:  nil,
			},
			expected: map[nat.Port][]nat.PortBinding{},
		},
		{
			name: "parsed port map without config",
			arg: arg{
				configPortMap: nil,
				parsedPortMap: map[nat.Port][]nat.PortBinding{
					"80/tcp": {{HostIP: "", HostPort: ""}},
				},
				exposedPorts: nil,
			},
			expected: map[nat.Port][]nat.PortBinding{
				"80/tcp": {{HostIP: "", HostPort: ""}},
			},
		},
		{
			name: "parsed and configured but not exposed",
			arg: arg{
				configPortMap: map[nat.Port][]nat.PortBinding{
					"80/tcp": {{HostIP: "1", HostPort: "2"}},
				},
				parsedPortMap: map[nat.Port][]nat.PortBinding{
					"80/tcp": {{HostIP: "", HostPort: ""}},
				},
				exposedPorts: nil,
			},
			expected: map[nat.Port][]nat.PortBinding{
				"80/tcp": {{HostIP: "", HostPort: ""}},
			},
		},
		{
			name: "merge both parsed and config",
			arg: arg{
				configPortMap: map[nat.Port][]nat.PortBinding{
					"60/tcp": {{HostIP: "1", HostPort: "2"}},
					"70/tcp": {{HostIP: "1", HostPort: "2"}},
					"80/tcp": {{HostIP: "1", HostPort: "2"}},
				},
				parsedPortMap: map[nat.Port][]nat.PortBinding{
					"80/tcp": {{HostIP: "", HostPort: ""}},
					"90/tcp": {{HostIP: "", HostPort: ""}},
				},
				exposedPorts: []string{"70", "80/tcp"},
			},
			expected: map[nat.Port][]nat.PortBinding{
				"70/tcp": {{HostIP: "1", HostPort: "2"}},
				"80/tcp": {{HostIP: "1", HostPort: "2"}},
				"90/tcp": {{HostIP: "", HostPort: ""}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := mergePortBindings(c.arg.configPortMap, c.arg.parsedPortMap, c.arg.exposedPorts)
			assert.Equal(t, c.expected, res)
		})
	}
}

func TestPortMappingCheck(t *testing.T) {
	makePortMap := func(ports ...string) nat.PortMap {
		out := make(nat.PortMap)
		for _, port := range ports {
			// We don't care about the actual binding in this test
			out[nat.Port(port)] = nil
		}
		return out
	}

	tests := map[string]struct {
		exposedAndMappedPorts nat.PortMap
		exposedPorts          []string
		expectError           bool
	}{
		"no-protocol": {
			exposedAndMappedPorts: makePortMap("1024/tcp"),
			exposedPorts:          []string{"1024"},
		},
		"protocol": {
			exposedAndMappedPorts: makePortMap("1024/tcp"),
			exposedPorts:          []string{"1024/tcp"},
		},
		"protocol-target-port": {
			exposedAndMappedPorts: makePortMap("1024/tcp"),
			exposedPorts:          []string{"1024:1024/tcp"},
		},
		"target-port": {
			exposedAndMappedPorts: makePortMap("1024/tcp"),
			exposedPorts:          []string{"1024:1024"},
		},
		"multiple-ports": {
			exposedAndMappedPorts: makePortMap("1024/tcp", "1025/tcp", "1026/tcp"),
			exposedPorts:          []string{"1024", "25:1025/tcp", "1026:1026"},
		},
		"only-ipv4": {
			exposedAndMappedPorts: makePortMap("1024/tcp"),
			exposedPorts:          []string{"0.0.0.0::1024/tcp"},
		},
		"no-mapped-ports": {
			exposedAndMappedPorts: makePortMap(),
			exposedPorts:          []string{"1024"},
			expectError:           true,
		},
		"wrong-mapped-port": {
			exposedAndMappedPorts: makePortMap("1023/tcp"),
			exposedPorts:          []string{"1024"},
			expectError:           true,
		},
		"subset-mapped-ports": {
			exposedAndMappedPorts: makePortMap("1024/tcp", "1025/tcp"),
			exposedPorts:          []string{"1024", "1025", "1026"},
			expectError:           true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkPortsMapped(tt.exposedAndMappedPorts, tt.exposedPorts)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestLifecycleHooks(t *testing.T) {
	tests := []struct {
		name  string
		reuse bool
	}{
		{
			name:  "GenericContainer",
			reuse: false,
		},
		{
			name:  "ReuseContainer",
			reuse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prints := []string{}
			ctx := context.Background()
			// reqWithLifecycleHooks {
			req := ContainerRequest{
				Image: nginxAlpineImage,
				LifecycleHooks: []ContainerLifecycleHooks{
					{
						PreCreates: []ContainerRequestHook{
							func(ctx context.Context, req ContainerRequest) error {
								prints = append(prints, "pre-create hook 1")
								return nil
							},
							func(ctx context.Context, req ContainerRequest) error {
								prints = append(prints, "pre-create hook 2")
								return nil
							},
						},
						PostCreates: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, "post-create hook 1")
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, "post-create hook 2")
								return nil
							},
						},
						PreStarts: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, "pre-start hook 1")
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, "pre-start hook 2")
								return nil
							},
						},
						PostStarts: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, "post-start hook 1")
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, "post-start hook 2")
								return nil
							},
						},
						PostReadies: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, "post-ready hook 1")
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, "post-ready hook 2")
								return nil
							},
						},
						PreStops: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, "pre-stop hook 1")
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, "pre-stop hook 2")
								return nil
							},
						},
						PostStops: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, "post-stop hook 1")
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, "post-stop hook 2")
								return nil
							},
						},
						PreTerminates: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, "pre-terminate hook 1")
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, "pre-terminate hook 2")
								return nil
							},
						},
						PostTerminates: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, "post-terminate hook 1")
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, "post-terminate hook 2")
								return nil
							},
						},
					},
				},
			}
			// }

			if tt.reuse {
				req.Name = "reuse-container"
			}

			c, err := GenericContainer(ctx, GenericContainerRequest{
				ContainerRequest: req,
				Reuse:            tt.reuse,
				Started:          true,
			})
			CleanupContainer(t, c)
			require.NoError(t, err)
			require.NotNil(t, c)

			duration := 1 * time.Second
			err = c.Stop(ctx, &duration)
			require.NoError(t, err)

			err = c.Start(ctx)
			require.NoError(t, err)

			err = c.Terminate(ctx)
			require.NoError(t, err)

			lifecycleHooksIsHonouredFn(t, prints)
		})
	}
}

// customLoggerImplementation {
type inMemoryLogger struct {
	data []string
}

func (l *inMemoryLogger) Printf(format string, args ...interface{}) {
	l.data = append(l.data, fmt.Sprintf(format, args...))
}

// }

func TestLifecycleHooks_WithDefaultLogger(t *testing.T) {
	ctx := context.Background()

	// reqWithDefaultLogginHook {
	dl := inMemoryLogger{}

	req := ContainerRequest{
		Image: nginxAlpineImage,
		LifecycleHooks: []ContainerLifecycleHooks{
			DefaultLoggingHook(&dl),
		},
	}
	// }

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	CleanupContainer(t, c)
	require.NoError(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.NoError(t, err)

	err = c.Start(ctx)
	require.NoError(t, err)

	err = c.Terminate(ctx)
	require.NoError(t, err)

	// Includes two additional entries for stop when terminate is called.
	require.Len(t, dl.data, 14)
}

func TestCombineLifecycleHooks(t *testing.T) {
	prints := []string{}

	preCreateFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, req ContainerRequest) error {
		return func(ctx context.Context, _ ContainerRequest) error {
			prints = append(prints, fmt.Sprintf("[%s] pre-%s hook %d.%d", prefix, hook, lifecycleID, hookID))
			return nil
		}
	}
	hookFunc := func(prefix string, hookType string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c Container) error {
		return func(ctx context.Context, _ Container) error {
			prints = append(prints, fmt.Sprintf("[%s] %s-%s hook %d.%d", prefix, hookType, hook, lifecycleID, hookID))
			return nil
		}
	}
	preFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c Container) error {
		return hookFunc(prefix, "pre", hook, lifecycleID, hookID)
	}
	postFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c Container) error {
		return hookFunc(prefix, "post", hook, lifecycleID, hookID)
	}

	lifecycleHookFunc := func(prefix string, lifecycleID int) ContainerLifecycleHooks {
		return ContainerLifecycleHooks{
			PreCreates:     []ContainerRequestHook{preCreateFunc(prefix, "create", lifecycleID, 1), preCreateFunc(prefix, "create", lifecycleID, 2)},
			PostCreates:    []ContainerHook{postFunc(prefix, "create", lifecycleID, 1), postFunc(prefix, "create", lifecycleID, 2)},
			PreStarts:      []ContainerHook{preFunc(prefix, "start", lifecycleID, 1), preFunc(prefix, "start", lifecycleID, 2)},
			PostStarts:     []ContainerHook{postFunc(prefix, "start", lifecycleID, 1), postFunc(prefix, "start", lifecycleID, 2)},
			PostReadies:    []ContainerHook{postFunc(prefix, "ready", lifecycleID, 1), postFunc(prefix, "ready", lifecycleID, 2)},
			PreStops:       []ContainerHook{preFunc(prefix, "stop", lifecycleID, 1), preFunc(prefix, "stop", lifecycleID, 2)},
			PostStops:      []ContainerHook{postFunc(prefix, "stop", lifecycleID, 1), postFunc(prefix, "stop", lifecycleID, 2)},
			PreTerminates:  []ContainerHook{preFunc(prefix, "terminate", lifecycleID, 1), preFunc(prefix, "terminate", lifecycleID, 2)},
			PostTerminates: []ContainerHook{postFunc(prefix, "terminate", lifecycleID, 1), postFunc(prefix, "terminate", lifecycleID, 2)},
		}
	}

	defaultHooks := []ContainerLifecycleHooks{lifecycleHookFunc("default", 1), lifecycleHookFunc("default", 2)}
	userDefinedHooks := []ContainerLifecycleHooks{lifecycleHookFunc("user-defined", 1), lifecycleHookFunc("user-defined", 2), lifecycleHookFunc("user-defined", 3)}

	hooks := combineContainerHooks(defaultHooks, userDefinedHooks)

	// call all the hooks in the right order, honouring the lifecycle

	req := ContainerRequest{}
	err := hooks.Creating(context.Background())(req)
	require.NoError(t, err)

	c := &DockerContainer{}

	err = hooks.Created(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Starting(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Started(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Readied(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Stopping(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Stopped(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Terminating(context.Background())(c)
	require.NoError(t, err)
	err = hooks.Terminated(context.Background())(c)
	require.NoError(t, err)

	// assertions

	// There are 2 default container lifecycle hooks and 3 user-defined container lifecycle hooks.
	// Each lifecycle hook has 2 pre-create hooks and 2 post-create hooks.
	// That results in 16 hooks per lifecycle (8 defaults + 12 user-defined = 20)

	// There are 5 lifecycles (create, start, ready, stop, terminate),
	// but ready has only half of the hooks (it only has post), so we have 90 hooks in total.
	assert.Len(t, prints, 90)

	// The order of the hooks is:
	// - pre-X hooks: first default (2*2), then user-defined (3*2)
	// - post-X hooks: first user-defined (3*2), then default (2*2)

	for i := 0; i < 5; i++ {
		var hookType string
		// this is the particular order of execution for the hooks
		switch i {
		case 0:
			hookType = "create"
		case 1:
			hookType = "start"
		case 2:
			hookType = "ready"
		case 3:
			hookType = "stop"
		case 4:
			hookType = "terminate"
		}

		initialIndex := i * 20
		if i >= 2 {
			initialIndex -= 10
		}

		if hookType != "ready" {
			// default pre-hooks: 4 hooks
			assert.Equal(t, fmt.Sprintf("[default] pre-%s hook 1.1", hookType), prints[initialIndex])
			assert.Equal(t, fmt.Sprintf("[default] pre-%s hook 1.2", hookType), prints[initialIndex+1])
			assert.Equal(t, fmt.Sprintf("[default] pre-%s hook 2.1", hookType), prints[initialIndex+2])
			assert.Equal(t, fmt.Sprintf("[default] pre-%s hook 2.2", hookType), prints[initialIndex+3])

			// user-defined pre-hooks: 6 hooks
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 1.1", hookType), prints[initialIndex+4])
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 1.2", hookType), prints[initialIndex+5])
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 2.1", hookType), prints[initialIndex+6])
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 2.2", hookType), prints[initialIndex+7])
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 3.1", hookType), prints[initialIndex+8])
			assert.Equal(t, fmt.Sprintf("[user-defined] pre-%s hook 3.2", hookType), prints[initialIndex+9])
		}

		// user-defined post-hooks: 6 hooks
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 1.1", hookType), prints[initialIndex+10])
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 1.2", hookType), prints[initialIndex+11])
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 2.1", hookType), prints[initialIndex+12])
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 2.2", hookType), prints[initialIndex+13])
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 3.1", hookType), prints[initialIndex+14])
		assert.Equal(t, fmt.Sprintf("[user-defined] post-%s hook 3.2", hookType), prints[initialIndex+15])

		// default post-hooks: 4 hooks
		assert.Equal(t, fmt.Sprintf("[default] post-%s hook 1.1", hookType), prints[initialIndex+16])
		assert.Equal(t, fmt.Sprintf("[default] post-%s hook 1.2", hookType), prints[initialIndex+17])
		assert.Equal(t, fmt.Sprintf("[default] post-%s hook 2.1", hookType), prints[initialIndex+18])
		assert.Equal(t, fmt.Sprintf("[default] post-%s hook 2.2", hookType), prints[initialIndex+19])
	}
}

func TestLifecycleHooks_WithMultipleHooks(t *testing.T) {
	ctx := context.Background()

	dl := inMemoryLogger{}

	req := ContainerRequest{
		Image: nginxAlpineImage,
		LifecycleHooks: []ContainerLifecycleHooks{
			DefaultLoggingHook(&dl),
			DefaultLoggingHook(&dl),
		},
	}

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	CleanupContainer(t, c)
	require.NoError(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.NoError(t, err)

	err = c.Start(ctx)
	require.NoError(t, err)

	err = c.Terminate(ctx)
	require.NoError(t, err)

	// Includes four additional entries for stop (twice) when terminate is called.
	require.Len(t, dl.data, 28)
}

type linesTestLogger struct {
	data []string
}

func (l *linesTestLogger) Printf(format string, args ...interface{}) {
	l.data = append(l.data, fmt.Sprintf(format, args...))
}

func TestPrintContainerLogsOnError(t *testing.T) {
	ctx := context.Background()

	req := ContainerRequest{
		Image:      "docker.io/alpine",
		Cmd:        []string{"echo", "-n", "I am expecting this"},
		WaitingFor: wait.ForLog("I was expecting that").WithStartupTimeout(5 * time.Second),
	}

	arrayOfLinesLogger := linesTestLogger{
		data: []string{},
	}

	ctr, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Logger:           &arrayOfLinesLogger,
		Started:          true,
	})
	CleanupContainer(t, ctr)
	// it should fail because the waiting for condition is not met
	if err == nil {
		t.Fatal(err)
	}

	containerLogs, err := ctr.Logs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer containerLogs.Close()

	// read container logs line by line, checking that each line is present in the stdout
	rd := bufio.NewReader(containerLogs)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}

			t.Fatal("Read Error:", err)
		}

		// the last line of the array should contain the line of interest,
		// but we are checking all the lines to make sure that is present
		found := false
		for _, l := range arrayOfLinesLogger.data {
			if strings.Contains(l, line) {
				found = true
				break
			}
		}
		assert.True(t, found, "container log line not found in the output of the logger: %s", line)
	}
}

func lifecycleHooksIsHonouredFn(t *testing.T, prints []string) {
	t.Helper()

	expects := []string{
		"pre-create hook 1",
		"pre-create hook 2",
		"post-create hook 1",
		"post-create hook 2",
		"pre-start hook 1",
		"pre-start hook 2",
		"post-start hook 1",
		"post-start hook 2",
		"post-ready hook 1",
		"post-ready hook 2",
		"pre-stop hook 1",
		"pre-stop hook 2",
		"post-stop hook 1",
		"post-stop hook 2",
		"pre-start hook 1",
		"pre-start hook 2",
		"post-start hook 1",
		"post-start hook 2",
		"post-ready hook 1",
		"post-ready hook 2",
		// Terminate currently calls stop to ensure that child containers are stopped.
		"pre-stop hook 1",
		"pre-stop hook 2",
		"post-stop hook 1",
		"post-stop hook 2",
		"pre-terminate hook 1",
		"pre-terminate hook 2",
		"post-terminate hook 1",
		"post-terminate hook 2",
	}

	require.Equal(t, expects, prints)
}
