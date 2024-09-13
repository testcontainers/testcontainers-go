package testcontainers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
	tcmount "github.com/testcontainers/testcontainers-go/mount"
)

func TestCreateContainerWithVolume(t *testing.T) {
	volumeName := "test-volume"
	// volumeMounts {
	req := testcontainers.Request{
		Image: "alpine",
		Mounts: tcmount.ContainerMounts{
			{
				Source: tcmount.GenericVolumeSource{
					Name: volumeName,
				},
				Target: "/data",
			},
		},
		Started: true,
	}
	// }

	ctx := context.Background()
	c, err := testcontainers.Run(ctx, req)
	testcontainers.CleanupContainer(t, c, testcontainers.RemoveVolumes(volumeName))
	require.NoError(t, err)

	// Check if volume is created
	client, err := core.NewClient(ctx)
	require.NoError(t, err)
	defer client.Close()

	volume, err := client.VolumeInspect(ctx, "test-volume")
	require.NoError(t, err)
	assert.Equal(t, "test-volume", volume.Name)
}

func TestMountsReceiveRyukLabels(t *testing.T) {
	volumeName := "app-data"
	req := testcontainers.Request{
		Image: "alpine",
		Mounts: tcmount.ContainerMounts{
			{
				Source: tcmount.GenericVolumeSource{
					Name: volumeName,
				},
				Target: "/data",
			},
		},
		Started: true,
	}

	ctx := context.Background()
	client, err := core.NewClient(ctx)
	require.NoError(t, err)
	defer client.Close()

	// Ensure the volume is removed before creating the container
	// otherwise the volume will be reused and the labels won't be set.
	err = client.VolumeRemove(ctx, volumeName, true)
	require.NoError(t, err)

	c, err := testcontainers.Run(ctx, req)
	testcontainers.CleanupContainer(t, c, testcontainers.RemoveVolumes(volumeName))
	require.NoError(t, err)

	// Check if volume is created with the expected labels
	volume, err := client.VolumeInspect(ctx, volumeName)
	require.NoError(t, err)
	require.Equal(t, core.DefaultLabels(core.SessionID()), volume.Labels)
}
