package testcontainers

import (
	"testing"

	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/assert"
)

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
