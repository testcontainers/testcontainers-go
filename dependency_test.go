package testcontainers

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
	"strings"
	"testing"
	"time"
)

func Test_ContainerDependency(t *testing.T) {
	type TestCase struct {
		name                string
		configureDependants func(ctx context.Context, t *testing.T) []ContainerDependency
		containerRequest    ContainerRequest
		expectedEnv         []string
		expectedError       string
	}
	testCases := []TestCase{
		{
			name: "dependency's dns name is passed as an environment variable to parent container",
			configureDependants: func(ctx context.Context, t *testing.T) []ContainerDependency {
				nginxReq := ContainerRequest{
					Image:        nginxAlpineImage,
					ExposedPorts: []string{nginxDefaultPort},
					WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
				}

				terminateFn := func(container Container) {
					terminateContainerOnEnd(t, ctx, container)
				}

				return []ContainerDependency{
					NewContainerDependency(nginxReq, "FIRST_DEPENDENCY").WithCallback(terminateFn),
					NewContainerDependency(nginxReq, "SECOND_DEPENDENCY").WithCallback(terminateFn),
				}
			},
			containerRequest: ContainerRequest{
				Image:      nginxAlpineImage,
				Entrypoint: []string{"tail", "-f", "/dev/null"},
			},
			expectedEnv: []string{"FIRST_DEPENDENCY", "SECOND_DEPENDENCY"},
		},
		{
			name: "container fails to start when dependency fails to start",
			configureDependants: func(ctx context.Context, t *testing.T) []ContainerDependency {
				badReq := ContainerRequest{
					Image:        "bad image name",
					ExposedPorts: []string{"80/tcp"},
				}

				return []ContainerDependency{
					NewContainerDependency(badReq, "FIRST_DEPENDENCY"),
				}
			},
			containerRequest: ContainerRequest{
				Image:      "curlimages/curl:8.7.1",
				Entrypoint: []string{"tail", "-f", "/dev/null"},
			},
			expectedError: "failed to create container",
		},
		{
			name: "fails to start dependency when key is empty",
			configureDependants: func(ctx context.Context, t *testing.T) []ContainerDependency {
				nginxReq := ContainerRequest{
					Image:        nginxAlpineImage,
					ExposedPorts: []string{nginxDefaultPort},
					WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
				}

				dependency := NewContainerDependency(nginxReq, "")
				return []ContainerDependency{dependency}
			},
			containerRequest: ContainerRequest{
				Image:      nginxAlpineImage,
				Entrypoint: []string{"tail", "-f", "/dev/null"},
			},
			expectedError: "cannot create dependency with empty environment key",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			dependantContainers := tc.configureDependants(ctx, t)
			tc.containerRequest.DependsOn = dependantContainers
			c, err := GenericContainer(ctx, GenericContainerRequest{
				ContainerRequest: tc.containerRequest,
				Started:          true,
			})

			if tc.expectedError != "" {
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
				require.True(t, c.IsRunning())

				inspection, err := c.(*DockerContainer).inspectContainer(ctx)
				require.NoError(t, err)

				for _, requiredEnv := range tc.expectedEnv {
					envVar := ""
					for _, env := range inspection.Config.Env {
						splitEnv := strings.Split(env, "=")
						if splitEnv[0] == requiredEnv {
							envVar = splitEnv[1]
							break
						}
					}
					require.NotEmpty(t, envVar)

					// Container.Exec cannot handle environment variables in the command.
					// As a workaround, env vars are fetched from the container and used as a substitute for the command.
					// In real scenarios, the running container can reference the env variables as normal.
					exitCode, _, err := c.Exec(ctx, []string{"curl", fmt.Sprintf("http://%v", envVar)})
					require.NoError(t, err)
					require.Equal(t, 0, exitCode)
				}
			}
			terminateContainerOnEnd(t, ctx, c)
		})
	}
}

func Test_ContainerDependency_CallbackFunc(t *testing.T) {
	ctx := context.Background()

	nginxReq := ContainerRequest{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
	}

	dependencyContainer := make(chan Container, 1)
	req := ContainerRequest{
		Image:      nginxAlpineImage,
		Entrypoint: []string{"tail", "-f", "/dev/null"},
		DependsOn: []ContainerDependency{
			NewContainerDependency(nginxReq, "MY_DEPENDENCY").
				WithCallback(func(c Container) {
					terminateContainerOnEnd(t, ctx, c)
					dependencyContainer <- c
				}),
		},
	}

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c)

	select {
	case <-time.After(3 * time.Second):
		t.Fatalf("dependency container callback was not called")
	case dependency := <-dependencyContainer:
		require.NotNil(t, dependency)
		require.True(t, dependency.IsRunning())
	}
}

func Test_ContainerDependency_ReuseRunningContainer(t *testing.T) {
	ctx := context.Background()

	nginxReq := ContainerRequest{
		Image:        nginxAlpineImage,
		Name:         "my-nginx-container",
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
	}
	depContainer, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: nginxReq,
		Started:          true,
	})
	require.NoError(t, err)
	require.True(t, depContainer.IsRunning())
	terminateContainerOnEnd(t, ctx, depContainer)

	dependencyContainer := make(chan Container, 1)
	req := ContainerRequest{
		Image:      nginxAlpineImage,
		Entrypoint: []string{"tail", "-f", "/dev/null"},
		DependsOn: []ContainerDependency{
			NewContainerDependency(nginxReq, "MY_DEPENDENCY").
				WithCallback(func(c Container) {
					dependencyContainer <- c
				}),
		},
	}
	c, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	require.True(t, c.IsRunning())
	terminateContainerOnEnd(t, ctx, c)

	select {
	case <-time.After(3 * time.Second):
		t.Fatalf("dependency container callback was not called")
	case dependency := <-dependencyContainer:
		require.Equal(t, depContainer.GetContainerID(), dependency.GetContainerID())
	}
}

func Test_ContainerDependency_KeepAlive(t *testing.T) {
	ctx := context.Background()

	nginxReq := ContainerRequest{
		Image:        nginxAlpineImage,
		Name:         "my-nginx-container",
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
	}

	testCases := []struct {
		name          string
		keepAlive     bool
		expectRunning bool
	}{
		{"dependency is terminated when KeepAlive is set to false", false, false},
		{"dependency is still running when KeepAlive is set to true", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var depContainer Container
			req := ContainerRequest{
				Image:      nginxAlpineImage,
				Entrypoint: []string{"tail", "-f", "/dev/null"},
				DependsOn: []ContainerDependency{
					NewContainerDependency(nginxReq, "MY_DEPENDENCY").
						WithCallback(func(c Container) {
							depContainer = c
						}).
						WithKeepAlive(tc.keepAlive),
				},
			}

			c, err := GenericContainer(ctx, GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			require.NoError(t, err)
			require.True(t, c.IsRunning())
			require.True(t, depContainer.IsRunning())
			if tc.keepAlive {
				terminateContainerOnEnd(t, ctx, depContainer)
			}

			err = c.Terminate(ctx)
			require.NoError(t, err)

			// Check the expected state of the dependency after the parent container is terminated.
			require.Equal(t, tc.expectRunning, depContainer.IsRunning())
		})
	}
}
