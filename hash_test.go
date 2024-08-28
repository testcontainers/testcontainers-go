package testcontainers

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestHashContainerRequest(t *testing.T) {
	req := ContainerRequest{
		Image: "nginx",
		Env:   map[string]string{"a": "b"},
		FromDockerfile: FromDockerfile{
			BuildOptionsModifier: func(options *types.ImageBuildOptions) {},
		},
		ExposedPorts:      []string{"80/tcp"},
		Privileged:        false,
		ImageSubstitutors: []ImageSubstitutor{newPrependHubRegistry("localhost:5000")},
		LifecycleHooks: []ContainerLifecycleHooks{
			{
				PreStarts: []ContainerHook{
					func(ctx context.Context, c Container) error {
						return nil
					},
				},
			},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			// NOOP
		},
		WaitingFor: wait.ForLog("nginx: ready"),
	}

	hash1, err := core.Hash(req)
	require.NoError(t, err)
	require.NotEqual(t, 0, hash1)

	hash2, err := core.Hash(req)
	require.NoError(t, err)
	require.NotEqual(t, 0, hash2)

	require.Equal(t, hash1, hash2)
}
