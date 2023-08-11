package testcontainers

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
)

func TestBindMount(t *testing.T) {
	t.Parallel()

	dockerSocket := testcontainersdocker.ExtractDockerSocket(context.Background())
	t.Log("Docker Socket Path: ", dockerSocket)

	type args struct {
		hostPath    string
		mountTarget ContainerMountTarget
	}
	tests := []struct {
		name string
		args args
		want ContainerMount
	}{
		{
			name: dockerSocket + ":" + dockerSocket,
			args: args{hostPath: dockerSocket, mountTarget: "/var/run/docker.sock"},
			want: ContainerMount{Source: GenericBindMountSource{HostPath: dockerSocket}, Target: "/var/run/docker.sock"},
		},
		{
			name: "/var/lib/app/data:/data",
			args: args{hostPath: "/var/lib/app/data", mountTarget: "/data"},
			want: ContainerMount{Source: GenericBindMountSource{HostPath: "/var/lib/app/data"}, Target: "/data"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, BindMount(tt.args.hostPath, tt.args.mountTarget), "BindMount(%v, %v)", tt.args.hostPath, tt.args.mountTarget)
		})
	}
}

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
			name:   "Single bind mount",
			mounts: ContainerMounts{{Source: GenericBindMountSource{HostPath: "/var/lib/app/data"}, Target: "/data"}},
			want: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/var/lib/app/data",
					Target: "/data",
				},
			},
		},
		{
			name:   "Single bind mount - read-only",
			mounts: ContainerMounts{{Source: GenericBindMountSource{HostPath: "/var/lib/app/data"}, Target: "/data", ReadOnly: true}},
			want: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   "/var/lib/app/data",
					Target:   "/data",
					ReadOnly: true,
				},
			},
		},
		{
			name: "Single bind mount - with options",
			mounts: ContainerMounts{
				{
					Source: DockerBindMountSource{
						HostPath: "/var/lib/app/data",
						BindOptions: &mount.BindOptions{
							Propagation: mount.PropagationPrivate,
						},
					},
					Target: "/data",
				},
			},
			want: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/var/lib/app/data",
					Target: "/data",
					BindOptions: &mount.BindOptions{
						Propagation: mount.PropagationPrivate,
					},
				},
			},
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
	client, err := NewDockerClient()
	assert.NoError(t, err)
	defer client.Close()

	volume, err := client.VolumeInspect(ctx, "test-volume")
	assert.NoError(t, err)
	assert.Equal(t, "test-volume", volume.Name)
}
