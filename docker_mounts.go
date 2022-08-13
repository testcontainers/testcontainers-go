package testcontainers

import "github.com/docker/docker/api/types/mount"

var (
	mountTypeMapping = map[MountType]mount.Type{
		MountTypeBind:   mount.TypeBind,
		MountTypeVolume: mount.TypeVolume,
		MountTypeTmpfs:  mount.TypeTmpfs,
		MountTypePipe:   mount.TypeNamedPipe,
	}
)

// BindMounter can optionally be implemented by mount sources
// to support advanced scenarios based on mount.BindOptions
type BindMounter interface {
	GetBindOptions() *mount.BindOptions
}

// VolumeMounter can optionally be implemented by mount sources
// to support advanced scenarios based on mount.VolumeOptions
type VolumeMounter interface {
	GetVolumeOptions() *mount.VolumeOptions
}

// TmpfsMounter can optionally be implemented by mount sources
// to support advanced scenarios based on mount.TmpfsOptions
type TmpfsMounter interface {
	GetTmpfsOptions() *mount.TmpfsOptions
}

type DockerBindMountSource struct {
	*mount.BindOptions

	// HostPath is the path mounted into the container
	// the same host path might be mounted to multiple locations withing a single container
	HostPath string
}

func (s DockerBindMountSource) Source() string {
	return s.HostPath
}

func (DockerBindMountSource) Type() MountType {
	return MountTypeBind
}

func (s DockerBindMountSource) GetBindOptions() *mount.BindOptions {
	return s.BindOptions
}

type DockerVolumeMountSource struct {
	*mount.VolumeOptions

	// Name refers to the name of the volume to be mounted
	// the same volume might be mounted to multiple locations within a single container
	Name string
}

func (s DockerVolumeMountSource) Source() string {
	return s.Name
}

func (DockerVolumeMountSource) Type() MountType {
	return MountTypeVolume
}

func (s DockerVolumeMountSource) GetVolumeOptions() *mount.VolumeOptions {
	return s.VolumeOptions
}

type DockerTmpfsMountSource struct {
	GenericTmpfsMountSource
	*mount.TmpfsOptions
}

func (s DockerTmpfsMountSource) GetTmpfsOptions() *mount.TmpfsOptions {
	return s.TmpfsOptions
}
