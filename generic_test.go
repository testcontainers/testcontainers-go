package testcontainers

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	reusableContainerName = "my_test_reusable_container"
)

func TestGenericReusableContainer(t *testing.T) {
	ctx := context.Background()

	n1, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "nginx:1.17.6",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForListeningPort("80/tcp"),
			Name:         reusableContainerName,
		},
		Started: true,
	})
	require.NoError(t, err)
	require.True(t, n1.IsRunning())
	defer n1.Terminate(ctx)

	copiedFileName := "hello_copy.sh"
	err = n1.CopyFileToContainer(ctx, "./testresources/hello.sh", "/"+copiedFileName, 700)
	require.NoError(t, err)

	tests := []struct {
		name          string
		containerName string
		errMsg        string
		reuseOption   bool
	}{
		{
			name:        "reuse option with empty name",
			errMsg:      ErrReuseEmptyName.Error(),
			reuseOption: true,
		},
		{
			name:          "container already exists (reuse=false)",
			containerName: reusableContainerName,
			errMsg:        "is already in use by container",
			reuseOption:   false,
		},
		{
			name:          "success reusing",
			containerName: reusableContainerName,
			reuseOption:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			n2, err := GenericContainer(ctx, GenericContainerRequest{
				ContainerRequest: ContainerRequest{
					Image:        "nginx:1.17.6",
					ExposedPorts: []string{"80/tcp"},
					WaitingFor:   wait.ForListeningPort("80/tcp"),
					Name:         tc.containerName,
				},
				Started: true,
				Reuse:   tc.reuseOption,
			})
			if tc.errMsg == "" {
				c, _, err := n2.Exec(ctx, []string{"bash", copiedFileName})
				require.NoError(t, err)
				require.Zero(t, c)
			} else {
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tc.errMsg))
			}
		})
	}

}
