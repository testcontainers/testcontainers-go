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
	require.Nil(t, err)
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
		require.Nil(t, err)

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
		require.Nil(t, err)

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
		require.Nil(t, err)

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
		require.Nil(t, err)
		defer func() {
			err := net.Remove(ctx)
			if err != nil {
				t.Logf("failed to remove network %s: %s\n", networkName, err)
			}
		}()

		dockerNetwork, err := provider.GetNetwork(ctx, NetworkRequest{
			Name: networkName,
		})
		require.Nil(t, err)

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
		require.Nil(t, err)

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
		require.Nil(t, err)
		defer func() {
			err := net.Remove(ctx)
			if err != nil {
				t.Logf("failed to remove network %s: %s\n", networkName, err)
			}
		}()

		dockerNetwork, err := provider.GetNetwork(ctx, NetworkRequest{
			Name: networkName,
		})
		require.Nil(t, err)

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
		require.Nil(t, err)

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

	t.Run("Request contains exposed port modifiers", func(t *testing.T) {
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
		require.Nil(t, err)

		// assertions
		assert.Equal(t, inputHostConfig.PortBindings["80/tcp"][0].HostIP, "localhost")
		assert.Equal(t, inputHostConfig.PortBindings["80/tcp"][0].HostPort, "8080")
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
				exposedPorts: []string{"70", "80"},
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
								prints = append(prints, fmt.Sprintf("pre-create hook 1: %#v", req))
								return nil
							},
							func(ctx context.Context, req ContainerRequest) error {
								prints = append(prints, fmt.Sprintf("pre-create hook 2: %#v", req))
								return nil
							},
						},
						PostCreates: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("post-create hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("post-create hook 2: %#v", c))
								return nil
							},
						},
						PreStarts: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("pre-start hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("pre-start hook 2: %#v", c))
								return nil
							},
						},
						PostStarts: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("post-start hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("post-start hook 2: %#v", c))
								return nil
							},
						},
						PreStops: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("pre-stop hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("pre-stop hook 2: %#v", c))
								return nil
							},
						},
						PostStops: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("post-stop hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("post-stop hook 2: %#v", c))
								return nil
							},
						},
						PreTerminates: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("pre-terminate hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("pre-terminate hook 2: %#v", c))
								return nil
							},
						},
						PostTerminates: []ContainerHook{
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("post-terminate hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c Container) error {
								prints = append(prints, fmt.Sprintf("post-terminate hook 2: %#v", c))
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
			require.Nil(t, err)
			require.NotNil(t, c)

			duration := 1 * time.Second
			err = c.Stop(ctx, &duration)
			require.Nil(t, err)

			err = c.Start(ctx)
			require.Nil(t, err)

			err = c.Terminate(ctx)
			require.Nil(t, err)

			lifecycleHooksIsHonouredFn(t, ctx, c, prints)
		})
	}
}

type inMemoryLogger struct {
	data []string
}

func (l *inMemoryLogger) Printf(format string, args ...interface{}) {
	l.data = append(l.data, fmt.Sprintf(format, args...))
}

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
	require.Nil(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.Nil(t, err)

	err = c.Start(ctx)
	require.Nil(t, err)

	err = c.Terminate(ctx)
	require.Nil(t, err)

	require.Equal(t, 10, len(dl.data))
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
	require.Nil(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.Nil(t, err)

	err = c.Start(ctx)
	require.Nil(t, err)

	err = c.Terminate(ctx)
	require.Nil(t, err)

	require.Equal(t, 20, len(dl.data))
}

type linesTestLogger struct {
	data []string
}

func (l *linesTestLogger) Printf(format string, args ...interface{}) {
	l.data = append(l.data, fmt.Sprintf(format, args...))
}

func TestPrintContainerLogsOnError(t *testing.T) {
	ctx := context.Background()
	client, err := NewDockerClientWithOpts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	req := ContainerRequest{
		Image:      "docker.io/alpine",
		Cmd:        []string{"echo", "-n", "I am expecting this"},
		WaitingFor: wait.ForLog("I was expecting that").WithStartupTimeout(5 * time.Second),
	}

	arrayOfLinesLogger := linesTestLogger{
		data: []string{},
	}

	container, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Logger:           &arrayOfLinesLogger,
		Started:          true,
	})
	// it should fail because the waiting for condition is not met
	if err == nil {
		t.Fatal(err)
	}
	terminateContainerOnEnd(t, ctx, container)

	containerLogs, err := container.Logs(ctx)
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

func lifecycleHooksIsHonouredFn(t *testing.T, ctx context.Context, container Container, prints []string) {
	require.Equal(t, 20, len(prints))

	assert.True(t, strings.HasPrefix(prints[0], "pre-create hook 1: "))
	assert.True(t, strings.HasPrefix(prints[1], "pre-create hook 2: "))

	assert.True(t, strings.HasPrefix(prints[2], "post-create hook 1: "))
	assert.True(t, strings.HasPrefix(prints[3], "post-create hook 2: "))

	assert.True(t, strings.HasPrefix(prints[4], "pre-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[5], "pre-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[6], "post-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[7], "post-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[8], "pre-stop hook 1: "))
	assert.True(t, strings.HasPrefix(prints[9], "pre-stop hook 2: "))

	assert.True(t, strings.HasPrefix(prints[10], "post-stop hook 1: "))
	assert.True(t, strings.HasPrefix(prints[11], "post-stop hook 2: "))

	assert.True(t, strings.HasPrefix(prints[12], "pre-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[13], "pre-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[14], "post-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[15], "post-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[16], "pre-terminate hook 1: "))
	assert.True(t, strings.HasPrefix(prints[17], "pre-terminate hook 2: "))

	assert.True(t, strings.HasPrefix(prints[18], "post-terminate hook 1: "))
	assert.True(t, strings.HasPrefix(prints[19], "post-terminate hook 2: "))
}
