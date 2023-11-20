package testcontainers

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/assert"
)

func TestVolumeMount(t *testing.T) {
	t.Parallel()
	type args struct {
		volumeName  string
		mountTarget ContainerMountTarget
	}
	tests := []struct {
		name string
		args args
		want ContainerMount
	}{
		{
			name: "sample-data:/data",
			args: args{volumeName: "sample-data", mountTarget: "/data"},
			want: ContainerMount{Source: GenericVolumeMountSource{Name: "sample-data"}, Target: "/data"},
		},
		{
			name: "web:/var/nginx/html",
			args: args{volumeName: "web", mountTarget: "/var/nginx/html"},
			want: ContainerMount{Source: GenericVolumeMountSource{Name: "web"}, Target: "/var/nginx/html"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, VolumeMount(tt.args.volumeName, tt.args.mountTarget), "VolumeMount(%v, %v)", tt.args.volumeName, tt.args.mountTarget)
		})
	}
}

func TestContainerMounts_PrepareMounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mounts ContainerMounts
		want   []mount.Mount
	}{
		{
			name:   "Empty",
			mounts: nil,
			want:   make([]mount.Mount, 0),
		},
		{
			name:   "Single volume mount",
			mounts: ContainerMounts{{Source: GenericVolumeMountSource{Name: "app-data"}, Target: "/data"}},
			want: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: "app-data",
					Target: "/data",
				},
			},
		},
		{
			name:   "Single volume mount - read-only",
			mounts: ContainerMounts{{Source: GenericVolumeMountSource{Name: "app-data"}, Target: "/data", ReadOnly: true}},
			want: []mount.Mount{
				{
					Type:     mount.TypeVolume,
					Source:   "app-data",
					Target:   "/data",
					ReadOnly: true,
				},
			},
		},
		{
			name: "Single volume mount - with options",
			mounts: ContainerMounts{
				{
					Source: DockerVolumeMountSource{
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
						Labels: map[string]string{
							"hello": "world",
						},
					},
				},
			},
		},

		{
			name:   "Single tmpfs mount",
			mounts: ContainerMounts{{Source: GenericTmpfsMountSource{}, Target: "/data"}},
			want: []mount.Mount{
				{
					Type:   mount.TypeTmpfs,
					Target: "/data",
				},
			},
		},
		{
			name:   "Single volume mount - read-only",
			mounts: ContainerMounts{{Source: GenericTmpfsMountSource{}, Target: "/data", ReadOnly: true}},
			want: []mount.Mount{
				{
					Type:     mount.TypeTmpfs,
					Target:   "/data",
					ReadOnly: true,
				},
			},
		},
		{
			name: "Single volume mount - with options",
			mounts: ContainerMounts{
				{
					Source: DockerTmpfsMountSource{
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
			assert.Equalf(t, tt.want, mapToDockerMounts(tt.mounts), "PrepareMounts()")
		})
	}
}

func TestCreateContainerWithVolume(t *testing.T) {
	// volumeMounts {
	req := ContainerRequest{
		Image: "alpine",
		Mounts: ContainerMounts{
			{
				Source: GenericVolumeMountSource{
					Name: "test-volume",
				},
				Target: "/data",
			},
		},
	}
	// }

	ctx := context.Background()
	c, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)
	terminateContainerOnEnd(t, ctx, c)

	// Check if volume is created
	client, err := NewDockerClientWithOpts(ctx)
	assert.NoError(t, err)
	defer client.Close()

	volume, err := client.VolumeInspect(ctx, "test-volume")
	assert.NoError(t, err)
	assert.Equal(t, "test-volume", volume.Name)
}
