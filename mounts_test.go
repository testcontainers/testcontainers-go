package testcontainers_test

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestVolumeMount(t *testing.T) {
	t.Parallel()
	type args struct {
		volumeName  string
		mountTarget testcontainers.ContainerMountTarget
	}
	tests := []struct {
		name string
		args args
		want testcontainers.ContainerMount
	}{
		{
			name: "sample-data:/data",
			args: args{volumeName: "sample-data", mountTarget: "/data"},
			want: testcontainers.ContainerMount{Source: testcontainers.GenericVolumeMountSource{Name: "sample-data"}, Target: "/data"},
		},
		{
			name: "web:/var/nginx/html",
			args: args{volumeName: "web", mountTarget: "/var/nginx/html"},
			want: testcontainers.ContainerMount{Source: testcontainers.GenericVolumeMountSource{Name: "web"}, Target: "/var/nginx/html"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, testcontainers.VolumeMount(tt.args.volumeName, tt.args.mountTarget), "VolumeMount(%v, %v)", tt.args.volumeName, tt.args.mountTarget)
		})
	}
}

func TestContainerMounts_PrepareMounts(t *testing.T) {
	volumeOptions := &mount.VolumeOptions{
		Labels: testcontainers.GenericLabels(),
	}

	expectedLabels := testcontainers.GenericLabels()
	expectedLabels["hello"] = "world"

	t.Parallel()
	tests := []struct {
		name   string
		mounts testcontainers.ContainerMounts
		want   []mount.Mount
	}{
		{
			name:   "Empty",
			mounts: nil,
			want:   make([]mount.Mount, 0),
		},
		{
			name:   "Single volume mount",
			mounts: testcontainers.ContainerMounts{{Source: testcontainers.GenericVolumeMountSource{Name: "app-data"}, Target: "/data"}},
			want: []mount.Mount{
				{
					Type:          mount.TypeVolume,
					Source:        "app-data",
					Target:        "/data",
					VolumeOptions: volumeOptions,
				},
			},
		},
		{
			name:   "Single volume mount - read-only",
			mounts: testcontainers.ContainerMounts{{Source: testcontainers.GenericVolumeMountSource{Name: "app-data"}, Target: "/data", ReadOnly: true}},
			want: []mount.Mount{
				{
					Type:          mount.TypeVolume,
					Source:        "app-data",
					Target:        "/data",
					ReadOnly:      true,
					VolumeOptions: volumeOptions,
				},
			},
		},
		{
			name: "Single volume mount - with options",
			mounts: testcontainers.ContainerMounts{
				{
					Source: testcontainers.DockerVolumeMountSource{
						Name: "app-data",
						VolumeOptions: &mount.VolumeOptions{
							NoCopy: true,
							Labels: map[string]string{
								"hello": "world",
							},
						},
					},
					Target: "/data",
				},
			},
			want: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: "app-data",
					Target: "/data",
					VolumeOptions: &mount.VolumeOptions{
						NoCopy: true,
						Labels: expectedLabels,
					},
				},
			},
		},

		{
			name:   "Single tmpfs mount",
			mounts: testcontainers.ContainerMounts{{Source: testcontainers.GenericTmpfsMountSource{}, Target: "/data"}},
			want: []mount.Mount{
				{
					Type:   mount.TypeTmpfs,
					Target: "/data",
				},
			},
		},
		{
			name:   "Single volume mount - read-only",
			mounts: testcontainers.ContainerMounts{{Source: testcontainers.GenericTmpfsMountSource{}, Target: "/data", ReadOnly: true}},
			want: []mount.Mount{
				{
					Type:     mount.TypeTmpfs,
					Target:   "/data",
					ReadOnly: true,
				},
			},
		},
		{
			name: "Single tmpfs mount - with options",
			mounts: testcontainers.ContainerMounts{
				{
					Source: testcontainers.DockerTmpfsMountSource{
						TmpfsOptions: &mount.TmpfsOptions{
							SizeBytes: 50 * 1024 * 1024,
							Mode:      0o644,
						},
					},
					Target: "/data",
				},
			},
			want: []mount.Mount{
				{
					Type:   mount.TypeTmpfs,
					Target: "/data",
					TmpfsOptions: &mount.TmpfsOptions{
						SizeBytes: 50 * 1024 * 1024,
						Mode:      0o644,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, tt.mounts.PrepareMounts(), "PrepareMounts()")
		})
	}
}

func TestCreateContainerWithVolume(t *testing.T) {
	volumeName := "test-volume"
	// volumeMounts {
	req := testcontainers.ContainerRequest{
		Image: "alpine",
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericVolumeMountSource{
					Name: volumeName,
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
	testcontainers.CleanupContainer(t, c, testcontainers.RemoveVolumes(volumeName))
	require.NoError(t, err)

	// Check if volume is created
	client, err := testcontainers.NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer client.Close()

	volume, err := client.VolumeInspect(ctx, "test-volume")
	require.NoError(t, err)
	assert.Equal(t, "test-volume", volume.Name)
}

func TestMountsReceiveRyukLabels(t *testing.T) {
	volumeName := "app-data"
	req := testcontainers.ContainerRequest{
		Image: "alpine",
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericVolumeMountSource{
					Name: volumeName,
				},
				Target: "/data",
			},
		},
	}

	ctx := context.Background()
	client, err := testcontainers.NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer client.Close()

	// Ensure the volume is removed before creating the container
	// otherwise the volume will be reused and the labels won't be set.
	err = client.VolumeRemove(ctx, volumeName, true)
	require.NoError(t, err)

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	testcontainers.CleanupContainer(t, c, testcontainers.RemoveVolumes(volumeName))
	require.NoError(t, err)

	// Check if volume is created with the expected labels.
	volume, err := client.VolumeInspect(ctx, volumeName)
	require.NoError(t, err)
	require.Equal(t, testcontainers.GenericLabels(), volume.Labels)
}
