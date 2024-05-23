package testcontainers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcmount "github.com/testcontainers/testcontainers-go/mount"
)

func TestCreateContainerWithVolume(t *testing.T) {
	// volumeMounts {
	req := testcontainers.ContainerRequest{
		Image: "alpine",
		Mounts: tcmount.ContainerMounts{
			{
				Source: tcmount.GenericVolumeSource{
					Name: "test-volume",
				},
				Target: "/data",
			},
		},
	}
	// }

	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c)

	// Check if volume is created
	client, err := testcontainers.NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer client.Close()

	volume, err := client.VolumeInspect(ctx, "test-volume")
	require.NoError(t, err)
	assert.Equal(t, "test-volume", volume.Name)
}

func TestMountsReceiveRyukLabels(t *testing.T) {
	req := testcontainers.ContainerRequest{
		Image: "alpine",
		Mounts: tcmount.ContainerMounts{
			{
				Source: tcmount.GenericVolumeSource{
					Name: "app-data",
				},
				Target: "/data",
			},
		},
	}

	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c)

	// Check if volume is created with the expected labels
	client, err := testcontainers.NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer client.Close()

	volume, err := client.VolumeInspect(ctx, "app-data")
	require.NoError(t, err)
	assert.Equal(t, testcontainers.GenericLabels(), volume.Labels)
}
