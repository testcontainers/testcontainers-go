package testcontainers

import (
	"github.com/docker/docker/api/types/mount"
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

// ContainerMountSource is the base for all mount sources
type ContainerMountSource interface {
	// Source will be used as Source field in the final mount
	// this might either be a volume name, a host path or might be empty e.g. for Tmpfs
	Source() string

	// Type determines the final mount type
	// possible options are limited by the Docker API
	Type() mount.Type
}

// BindMountSource implements ContainerMountSource and represents a bind mount
// Optionally mount.BindOptions might be added for advanced scenarios
type BindMountSource struct {
	*mount.BindOptions

	// HostPath is the path mounted into the container
	// the same host path might be mounted to multiple locations withing a single container
	HostPath string
}

func (s BindMountSource) Source() string {
	return s.HostPath
}

func (BindMountSource) Type() mount.Type {
	return mount.TypeBind
}

func (s BindMountSource) GetBindOptions() *mount.BindOptions {
	return s.BindOptions
}

// VolumeMountSource implements ContainerMountSource and represents a volume mount
// Optionally mount.VolumeOptions might be added for advanced scenarios
type VolumeMountSource struct {
	*mount.VolumeOptions

	// Name refers to the name of the volume to be mounted
	// the same volume might be mounted to multiple locations within a single container
	Name string
}

func (s VolumeMountSource) Source() string {
	return s.Name
}

func (VolumeMountSource) Type() mount.Type {
	return mount.TypeVolume
}

func (s VolumeMountSource) GetVolumeOptions() *mount.VolumeOptions {
	return s.VolumeOptions
}

// TmpfsMountSource implements ContainerMountSource and represents a TmpFS mount
// Optionally mount.TmpfsOptions might be added for advanced scenarios
type TmpfsMountSource struct {
	*mount.TmpfsOptions
}

func (s TmpfsMountSource) Source() string {
	return ""
}

func (TmpfsMountSource) Type() mount.Type {
	return mount.TypeTmpfs
}

func (s TmpfsMountSource) GetTmpfsOptions() *mount.TmpfsOptions {
	return s.TmpfsOptions
}

// ContainerMountTarget represents the target path within a container where the mount will be available
// Note that mount targets must be unique. It's not supported to mount different sources to the same target.
type ContainerMountTarget string

func (t ContainerMountTarget) Target() string {
	return string(t)
}

// BindMount returns a new ContainerMount with a BindMountSource as source
// This is a convenience method to cover typical use cases.
func BindMount(hostPath string, mountTarget ContainerMountTarget) ContainerMount {
	return ContainerMount{
		Source: BindMountSource{HostPath: hostPath},
		Target: mountTarget,
	}
}

// VolumeMount returns a new ContainerMount with a VolumeMountSource as source
// This is a convenience method to cover typical use cases.
func VolumeMount(volumeName string, mountTarget ContainerMountTarget) ContainerMount {
	return ContainerMount{
		Source: VolumeMountSource{Name: volumeName},
		Target: mountTarget,
	}
}

// Mounts returns a ContainerMounts to support a more fluent API
func Mounts(mounts ...ContainerMount) ContainerMounts {
	return mounts
}

// ContainerMount models a mount into a container
type ContainerMount struct {
	// Source is typically either a BindMountSource or a VolumeMountSource
	Source ContainerMountSource
	// Target is the path where the mount should be mounted within the container
	Target ContainerMountTarget
	// ReadOnly determines if the mount should be read-only
	ReadOnly bool
}

// ContainerMounts represents a collection of mounts for a container
type ContainerMounts []ContainerMount

// PrepareMounts maps the given []ContainerMount to the corresponding
// []mount.Mount for further processing
func (m ContainerMounts) PrepareMounts() []mount.Mount {
	mounts := make([]mount.Mount, 0, len(m))

	for idx := range m {
		m := m[idx]
		containerMount := mount.Mount{
			Type:     m.Source.Type(),
			Source:   m.Source.Source(),
			ReadOnly: m.ReadOnly,
			Target:   m.Target.Target(),
		}

		switch typedMounter := m.Source.(type) {
		case BindMounter:
			containerMount.BindOptions = typedMounter.GetBindOptions()
		case VolumeMounter:
			containerMount.VolumeOptions = typedMounter.GetVolumeOptions()
		case TmpfsMounter:
			containerMount.TmpfsOptions = typedMounter.GetTmpfsOptions()
		}

		mounts = append(mounts, containerMount)
	}

	return mounts
}
