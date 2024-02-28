package testcontainers_test

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

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPreCreateModifierHook(t *testing.T) {
	ctx := context.Background()

	provider, err := testcontainers.NewDockerProvider()
	require.NoError(t, err)
	defer provider.Close()

	t.Run("No exposed ports", func(t *testing.T) {
		// reqWithModifiers {
		req := testcontainers.ContainerRequest{
			Image: nginxAlpineImage, // alpine image does expose port 80
			ConfigModifier: func(config *container.Config) {
				config.Env = []string{"a=b"}
			},
			Mounts: testcontainers.ContainerMounts{
				{
					Source: testcontainers.DockerVolumeMountSource{
						Name: "appdata",
						VolumeOptions: &mount.VolumeOptions{
							Labels: testcontainers.GenericLabels(),
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

		hooks := testcontainers.DefaultPreCreateHook(provider, inputConfig, inputHostConfig, inputNetworkingConfig)
		err = hooks.Creating(ctx)(req)
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
						Labels: testcontainers.GenericLabels(),
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
		req := testcontainers.ContainerRequest{
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

		hooks := testcontainers.DefaultPreCreateHook(provider, inputConfig, inputHostConfig, inputNetworkingConfig)
		err = hooks.Creating(ctx)(req)
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
		req := testcontainers.ContainerRequest{
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

		hooks := testcontainers.DefaultPreCreateHook(provider, inputConfig, inputHostConfig, inputNetworkingConfig)
		err = hooks.Creating(ctx)(req)
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
		net, err := provider.CreateNetwork(ctx, testcontainers.NetworkRequest{
			Name: networkName,
		})
		require.NoError(t, err)
		defer func() {
			err := net.Remove(ctx)
			if err != nil {
				t.Logf("failed to remove network %s: %s\n", networkName, err)
			}
		}()

		dockerNetwork, err := provider.GetNetwork(ctx, testcontainers.NetworkRequest{
			Name: networkName,
		})
		require.NoError(t, err)

		req := testcontainers.ContainerRequest{
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

		hooks := testcontainers.DefaultPreCreateHook(provider, inputConfig, inputHostConfig, inputNetworkingConfig)
		err = hooks.Creating(ctx)(req)
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
		net, err := provider.CreateNetwork(ctx, testcontainers.NetworkRequest{
			Name: networkName,
		})
		require.NoError(t, err)
		defer func() {
			err := net.Remove(ctx)
			if err != nil {
				t.Logf("failed to remove network %s: %s\n", networkName, err)
			}
		}()

		dockerNetwork, err := provider.GetNetwork(ctx, testcontainers.NetworkRequest{
			Name: networkName,
		})
		require.NoError(t, err)

		req := testcontainers.ContainerRequest{
			Image:    nginxAlpineImage, // alpine image does expose port 80
			Networks: []string{networkName, "bar"},
		}

		// define empty inputs to be overwritten by the pre create hook
		inputConfig := &container.Config{
			Image: req.Image,
		}
		inputHostConfig := &container.HostConfig{}
		inputNetworkingConfig := &network.NetworkingConfig{}

		hooks := testcontainers.DefaultPreCreateHook(provider, inputConfig, inputHostConfig, inputNetworkingConfig)
		err = hooks.Creating(ctx)(req)
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

	t.Run("Request contains exposed port modifiers", func(t *testing.T) {
		req := testcontainers.ContainerRequest{
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

		hooks := testcontainers.DefaultPreCreateHook(provider, inputConfig, inputHostConfig, inputNetworkingConfig)
		err = hooks.Creating(ctx)(req)
		require.NoError(t, err)

		// assertions
		assert.Equal(t, "localhost", inputHostConfig.PortBindings["80/tcp"][0].HostIP)
		assert.Equal(t, "8080", inputHostConfig.PortBindings["80/tcp"][0].HostPort)
	})
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
			req := testcontainers.ContainerRequest{
				Image: nginxAlpineImage,
				LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
					{
						PreCreates: []testcontainers.ContainerRequestHook{
							func(ctx context.Context, req testcontainers.ContainerRequest) error {
								prints = append(prints, fmt.Sprintf("pre-create hook 1: %#v", req))
								return nil
							},
							func(ctx context.Context, req testcontainers.ContainerRequest) error {
								prints = append(prints, fmt.Sprintf("pre-create hook 2: %#v", req))
								return nil
							},
						},
						PostCreates: []testcontainers.ContainerHook{
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("post-create hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("post-create hook 2: %#v", c))
								return nil
							},
						},
						PreStarts: []testcontainers.ContainerHook{
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("pre-start hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("pre-start hook 2: %#v", c))
								return nil
							},
						},
						PostStarts: []testcontainers.ContainerHook{
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("post-start hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("post-start hook 2: %#v", c))
								return nil
							},
						},
						PostReadies: []testcontainers.ContainerHook{
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("post-ready hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("post-ready hook 2: %#v", c))
								return nil
							},
						},
						PreStops: []testcontainers.ContainerHook{
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("pre-stop hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("pre-stop hook 2: %#v", c))
								return nil
							},
						},
						PostStops: []testcontainers.ContainerHook{
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("post-stop hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("post-stop hook 2: %#v", c))
								return nil
							},
						},
						PreTerminates: []testcontainers.ContainerHook{
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("pre-terminate hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("pre-terminate hook 2: %#v", c))
								return nil
							},
						},
						PostTerminates: []testcontainers.ContainerHook{
							func(ctx context.Context, c testcontainers.Container) error {
								prints = append(prints, fmt.Sprintf("post-terminate hook 1: %#v", c))
								return nil
							},
							func(ctx context.Context, c testcontainers.Container) error {
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

			c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Reuse:            tt.reuse,
				Started:          true,
			})
			require.NoError(t, err)
			require.NotNil(t, c)

			duration := 1 * time.Second
			err = c.Stop(ctx, &duration)
			require.NoError(t, err)

			err = c.Start(ctx)
			require.NoError(t, err)

			err = c.Terminate(ctx)
			require.NoError(t, err)

			lifecycleHooksIsHonouredFn(t, ctx, prints)
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

	req := testcontainers.ContainerRequest{
		Image: nginxAlpineImage,
		LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
			testcontainers.DefaultLoggingHook(&dl),
		},
	}
	// }

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.NoError(t, err)

	err = c.Start(ctx)
	require.NoError(t, err)

	err = c.Terminate(ctx)
	require.NoError(t, err)

	require.Len(t, dl.data, 12)
}

func TestCombineLifecycleHooks(t *testing.T) {
	prints := []string{}

	preCreateFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, req testcontainers.ContainerRequest) error {
		return func(ctx context.Context, _ testcontainers.ContainerRequest) error {
			prints = append(prints, fmt.Sprintf("[%s] pre-%s hook %d.%d", prefix, hook, lifecycleID, hookID))
			return nil
		}
	}
	hookFunc := func(prefix string, hookType string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c testcontainers.Container) error {
		return func(ctx context.Context, _ testcontainers.Container) error {
			prints = append(prints, fmt.Sprintf("[%s] %s-%s hook %d.%d", prefix, hookType, hook, lifecycleID, hookID))
			return nil
		}
	}
	preFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c testcontainers.Container) error {
		return hookFunc(prefix, "pre", hook, lifecycleID, hookID)
	}
	postFunc := func(prefix string, hook string, lifecycleID int, hookID int) func(ctx context.Context, c testcontainers.Container) error {
		return hookFunc(prefix, "post", hook, lifecycleID, hookID)
	}

	lifecycleHookFunc := func(prefix string, lifecycleID int) testcontainers.ContainerLifecycleHooks {
		return testcontainers.ContainerLifecycleHooks{
			PreCreates:     []testcontainers.ContainerRequestHook{preCreateFunc(prefix, "create", lifecycleID, 1), preCreateFunc(prefix, "create", lifecycleID, 2)},
			PostCreates:    []testcontainers.ContainerHook{postFunc(prefix, "create", lifecycleID, 1), postFunc(prefix, "create", lifecycleID, 2)},
			PreStarts:      []testcontainers.ContainerHook{preFunc(prefix, "start", lifecycleID, 1), preFunc(prefix, "start", lifecycleID, 2)},
			PostStarts:     []testcontainers.ContainerHook{postFunc(prefix, "start", lifecycleID, 1), postFunc(prefix, "start", lifecycleID, 2)},
			PostReadies:    []testcontainers.ContainerHook{postFunc(prefix, "ready", lifecycleID, 1), postFunc(prefix, "ready", lifecycleID, 2)},
			PreStops:       []testcontainers.ContainerHook{preFunc(prefix, "stop", lifecycleID, 1), preFunc(prefix, "stop", lifecycleID, 2)},
			PostStops:      []testcontainers.ContainerHook{postFunc(prefix, "stop", lifecycleID, 1), postFunc(prefix, "stop", lifecycleID, 2)},
			PreTerminates:  []testcontainers.ContainerHook{preFunc(prefix, "terminate", lifecycleID, 1), preFunc(prefix, "terminate", lifecycleID, 2)},
			PostTerminates: []testcontainers.ContainerHook{postFunc(prefix, "terminate", lifecycleID, 1), postFunc(prefix, "terminate", lifecycleID, 2)},
		}
	}

	defaultHooks := []testcontainers.ContainerLifecycleHooks{lifecycleHookFunc("default", 1), lifecycleHookFunc("default", 2)}
	userDefinedHooks := []testcontainers.ContainerLifecycleHooks{lifecycleHookFunc("user-defined", 1), lifecycleHookFunc("user-defined", 2), lifecycleHookFunc("user-defined", 3)}

	hooks := testcontainers.CombineContainerHooks(defaultHooks, userDefinedHooks)

	// call all the hooks in the right order, honouring the lifecycle

	req := testcontainers.ContainerRequest{}
	err := hooks.Creating(context.Background())(req)
	require.NoError(t, err)

	c := &testcontainers.DockerContainer{}

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

	req := testcontainers.ContainerRequest{
		Image: nginxAlpineImage,
		LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
			testcontainers.DefaultLoggingHook(&dl),
			testcontainers.DefaultLoggingHook(&dl),
		},
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	duration := 1 * time.Second
	err = c.Stop(ctx, &duration)
	require.NoError(t, err)

	err = c.Start(ctx)
	require.NoError(t, err)

	err = c.Terminate(ctx)
	require.NoError(t, err)

	require.Len(t, dl.data, 24)
}

type linesTestLogger struct {
	data []string
}

func (l *linesTestLogger) Printf(format string, args ...interface{}) {
	l.data = append(l.data, fmt.Sprintf(format, args...))
}

func TestPrintContainerLogsOnError(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:      "docker.io/alpine",
		Cmd:        []string{"echo", "-n", "I am expecting this"},
		WaitingFor: wait.ForLog("I was expecting that").WithStartupTimeout(5 * time.Second),
	}

	arrayOfLinesLogger := linesTestLogger{
		data: []string{},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
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

func lifecycleHooksIsHonouredFn(t *testing.T, ctx context.Context, prints []string) {
	require.Len(t, prints, 24)

	assert.True(t, strings.HasPrefix(prints[0], "pre-create hook 1: "))
	assert.True(t, strings.HasPrefix(prints[1], "pre-create hook 2: "))

	assert.True(t, strings.HasPrefix(prints[2], "post-create hook 1: "))
	assert.True(t, strings.HasPrefix(prints[3], "post-create hook 2: "))

	assert.True(t, strings.HasPrefix(prints[4], "pre-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[5], "pre-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[6], "post-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[7], "post-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[8], "post-ready hook 1: "))
	assert.True(t, strings.HasPrefix(prints[9], "post-ready hook 2: "))

	assert.True(t, strings.HasPrefix(prints[10], "pre-stop hook 1: "))
	assert.True(t, strings.HasPrefix(prints[11], "pre-stop hook 2: "))

	assert.True(t, strings.HasPrefix(prints[12], "post-stop hook 1: "))
	assert.True(t, strings.HasPrefix(prints[13], "post-stop hook 2: "))

	assert.True(t, strings.HasPrefix(prints[14], "pre-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[15], "pre-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[16], "post-start hook 1: "))
	assert.True(t, strings.HasPrefix(prints[17], "post-start hook 2: "))

	assert.True(t, strings.HasPrefix(prints[18], "post-ready hook 1: "))
	assert.True(t, strings.HasPrefix(prints[19], "post-ready hook 2: "))

	assert.True(t, strings.HasPrefix(prints[20], "pre-terminate hook 1: "))
	assert.True(t, strings.HasPrefix(prints[21], "pre-terminate hook 2: "))

	assert.True(t, strings.HasPrefix(prints[22], "post-terminate hook 1: "))
	assert.True(t, strings.HasPrefix(prints[23], "post-terminate hook 2: "))
}
