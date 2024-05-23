package mount_test

import (
	"testing"

	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go/internal/core"
	tcmount "github.com/testcontainers/testcontainers-go/mount"
)

func TestVolumeMount(t *testing.T) {
	t.Parallel()
	type args struct {
		volumeName  string
		mountTarget tcmount.ContainerTarget
	}
	tests := []struct {
		name string
		args args
		want tcmount.ContainerMount
	}{
		{
			name: "sample-data:/data",
			args: args{volumeName: "sample-data", mountTarget: "/data"},
			want: tcmount.ContainerMount{Source: tcmount.GenericVolumeSource{Name: "sample-data"}, Target: "/data"},
		},
		{
			name: "web:/var/nginx/html",
			args: args{volumeName: "web", mountTarget: "/var/nginx/html"},
			want: tcmount.ContainerMount{Source: tcmount.GenericVolumeSource{Name: "web"}, Target: "/var/nginx/html"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, tcmount.VolumeMount(tt.args.volumeName, tt.args.mountTarget), "VolumeMount(%v, %v)", tt.args.volumeName, tt.args.mountTarget)
		})
	}
}

func TestContainerMounts_PrepareMounts(t *testing.T) {
	volumeOptions := &mount.VolumeOptions{
		Labels: core.DefaultLabels(core.SessionID()),
	}

	expectedLabels := core.DefaultLabels(core.SessionID())
	expectedLabels["hello"] = "world"

	t.Parallel()
	tests := []struct {
		name   string
		mounts tcmount.ContainerMounts
		want   []mount.Mount
	}{
		{
			name:   "Empty",
			mounts: nil,
			want:   make([]mount.Mount, 0),
		},
		{
			name:   "Single volume mount",
			mounts: tcmount.ContainerMounts{{Source: tcmount.GenericVolumeSource{Name: "app-data"}, Target: "/data"}},
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
			mounts: tcmount.ContainerMounts{{Source: tcmount.GenericVolumeSource{Name: "app-data"}, Target: "/data", ReadOnly: true}},
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
			mounts: tcmount.ContainerMounts{
				{
					Source: tcmount.DockerVolumeSource{
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
			mounts: tcmount.ContainerMounts{{Source: tcmount.GenericTmpfsSource{}, Target: "/data"}},
			want: []mount.Mount{
				{
					Type:   mount.TypeTmpfs,
					Target: "/data",
				},
			},
		},
		{
			name:   "Single volume mount - read-only",
			mounts: tcmount.ContainerMounts{{Source: tcmount.GenericTmpfsSource{}, Target: "/data", ReadOnly: true}},
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
			mounts: tcmount.ContainerMounts{
				{
					Source: tcmount.DockerTmpfsSource{
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
			assert.Equalf(t, tt.want, tt.mounts.Prepare(), "Prepare()")
		})
	}
}
