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
	// volumeMounts {
	req := testcontainers.Request{
		Image: "alpine",
		Mounts: tcmount.ContainerMounts{
			{
				Source: tcmount.GenericVolumeSource{
					Name: "test-volume",
				},
				Target: "/data",
			},
		},
		Started: true,
	}
	// }

	ctx := context.Background()
	c, err := testcontainers.Run(ctx, req)
	require.NoError(t, err)
	testcontainers.TerminateContainerOnEnd(t, ctx, c)

	// Check if volume is created
	client, err := core.NewClient(ctx)
	require.NoError(t, err)
	defer client.Close()

	volume, err := client.VolumeInspect(ctx, "test-volume")
	require.NoError(t, err)
	assert.Equal(t, "test-volume", volume.Name)
}

func TestMountsReceiveRyukLabels(t *testing.T) {
	req := testcontainers.Request{
		Image: "alpine",
		Mounts: tcmount.ContainerMounts{
			{
				Source: tcmount.GenericVolumeSource{
					Name: "app-data",
				},
				Target: "/data",
			},
		},
		Started: true,
	}

	ctx := context.Background()
	c, err := testcontainers.Run(ctx, req)
	require.NoError(t, err)
	testcontainers.TerminateContainerOnEnd(t, ctx, c)

	// Check if volume is created with the expected labels
	client, err := core.NewClient(ctx)
	require.NoError(t, err)
	defer client.Close()

	volume, err := client.VolumeInspect(ctx, "app-data")
	require.NoError(t, err)
	assert.Equal(t, core.DefaultLabels(core.SessionID()), volume.Labels)
}
