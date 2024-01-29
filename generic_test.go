package testcontainers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	reusableContainerName = "my_test_reusable_container"
)

func TestGenericReusableContainer(t *testing.T) {
	ctx := context.Background()

	n1, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:        nginxAlpineImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
			Name:         reusableContainerName,
		},
		Started: true,
	})
	require.NoError(t, err)
	require.True(t, n1.IsRunning())
	terminateContainerOnEnd(t, ctx, n1)

	copiedFileName := "hello_copy.sh"
	err = n1.CopyFileToContainer(ctx, "./testdata/hello.sh", "/"+copiedFileName, 700)
	require.NoError(t, err)

	tests := []struct {
		name          string
		containerName string
		errorMatcher  func(err error) error
		reuseOption   bool
	}{
		{
			name: "reuse option with empty name",
			errorMatcher: func(err error) error {
				if errors.Is(err, ErrReuseEmptyName) {
					return nil
				}
				return err
			},
			reuseOption: true,
		},
		{
			name:          "container already exists (reuse=false)",
			containerName: reusableContainerName,
			errorMatcher: func(err error) error {
				if err == nil {
					return errors.New("expected error but got none")
				}
				return nil
			},
			reuseOption: false,
		},
		{
			name:          "success reusing",
			containerName: reusableContainerName,
			reuseOption:   true,
			errorMatcher: func(err error) error {
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			n2, err := GenericContainer(ctx, GenericContainerRequest{
				ProviderType: providerType,
				ContainerRequest: ContainerRequest{
					Image:        nginxAlpineImage,
					ExposedPorts: []string{nginxDefaultPort},
					WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
					Name:         tc.containerName,
				},
				Started: true,
				Reuse:   tc.reuseOption,
			})

			require.NoError(t, tc.errorMatcher(err))

			if err == nil {
				c, _, err := n2.Exec(ctx, []string{"/bin/ash", copiedFileName})
				require.NoError(t, err)
				require.Zero(t, c)
			}
		})
	}
}

func TestGenericContainerShouldReturnRefOnError(t *testing.T) {
	// In this test, we are going to cancel the context to exit the `wait.Strategy`.
	// We want to make sure that the GenericContainer call will still return a reference to the
	// created container, so that we can Destroy it.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType: providerType,
		ContainerRequest: ContainerRequest{
			Image:      nginxAlpineImage,
			WaitingFor: wait.ForLog("this string should not be present in the logs"),
		},
		Started: true,
	})
	require.Error(t, err)
	require.NotNil(t, c)
	terminateContainerOnEnd(t, context.Background(), c)
}
