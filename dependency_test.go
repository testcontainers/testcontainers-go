package testcontainers

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
	"strings"
	"testing"
)

func Test_ContainerDependsOn(t *testing.T) {
	type TestCase struct {
		name                string
		configureDependants func(ctx context.Context, t *testing.T) []*ContainerDependency
		containerRequest    ContainerRequest
		expectedEnv         []string
		expectedError       string
	}

	testCases := []TestCase{
		{
			name: "dependency is started and the dns name is passed as an environment variable",
			configureDependants: func(ctx context.Context, t *testing.T) []*ContainerDependency {
				nginxReq := ContainerRequest{
					Image:        nginxAlpineImage,
					ExposedPorts: []string{nginxDefaultPort},
					WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
				}

				return []*ContainerDependency{
					NewContainerDependency(nginxReq, "FIRST_DEPENDENCY"),
					NewContainerDependency(nginxReq, "SECOND_DEPENDENCY"),
				}
			},
			containerRequest: ContainerRequest{
				Image:      "curlimages/curl:8.7.1",
				Entrypoint: []string{"tail", "-f", "/dev/null"},
			},
			expectedEnv: []string{"FIRST_DEPENDENCY", "SECOND_DEPENDENCY"},
		},
		{
			name: "container fails to start when dependency fails to start",
			configureDependants: func(ctx context.Context, t *testing.T) []*ContainerDependency {
				badReq := ContainerRequest{
					Image:        "bad image name",
					ExposedPorts: []string{"80/tcp"},
				}

				return []*ContainerDependency{
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
			configureDependants: func(ctx context.Context, t *testing.T) []*ContainerDependency {
				nginxReq := ContainerRequest{
					Image:        nginxAlpineImage,
					ExposedPorts: []string{nginxDefaultPort},
					WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
				}

				dependency := NewContainerDependency(nginxReq, "")
				return []*ContainerDependency{dependency}
			},
			containerRequest: ContainerRequest{
				Image:      "curlimages/curl:8.7.1",
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
		})
	}
}
