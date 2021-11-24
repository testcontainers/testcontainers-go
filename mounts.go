package testcontainers

import (
	"fmt"

	"github.com/docker/docker/api/types/mount"
)

type BindMounter interface {
	GetBindOptions() *mount.BindOptions
}

type VolumeMounter interface {
	GetVolumeOptions() *mount.VolumeOptions
}

type TmpfsMounter interface {
	GetTmpfsOptions() *mount.TmpfsOptions
}

type ContainerMountSource interface {
	fmt.Stringer
	Type() mount.Type
}

type BindMountSource struct {
	*mount.BindOptions
	HostPath string
}

func (s BindMountSource) String() string {
	return s.HostPath
}

func (BindMountSource) Type() mount.Type {
	return mount.TypeBind
}
func (s BindMountSource) GetBindOptions() *mount.BindOptions {
	return s.BindOptions
}

type VolumeMountSource struct {
	*mount.VolumeOptions
	Name string
}

func (s VolumeMountSource) String() string {
	return s.Name
}

func (VolumeMountSource) Type() mount.Type {
	return mount.TypeVolume
}

func (s VolumeMountSource) GetVolumeOptions() *mount.VolumeOptions {
	return s.VolumeOptions
}

type TmpfsMountSource struct {
	*mount.TmpfsOptions
}

func (s TmpfsMountSource) String() string {
	return ""
}

func (TmpfsMountSource) Type() mount.Type {
	return mount.TypeTmpfs
}

func (s TmpfsMountSource) GetTmpfsOptions() *mount.TmpfsOptions {
	return s.TmpfsOptions
}

type ContainerMountTarget string

func (t ContainerMountTarget) String() string {
	return string(t)
}

func BindMount(hostPath string, mountTarget ContainerMountTarget) ContainerMount {
	return ContainerMount{
		Source: BindMountSource{HostPath: hostPath},
		Target: mountTarget,
	}
}

func VolumeMount(volumeName string, mountTarget ContainerMountTarget) ContainerMount {
	return ContainerMount{
		Source: VolumeMountSource{Name: volumeName},
		Target: mountTarget,
	}
}

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

type ContainerMounts []ContainerMount

func (m ContainerMounts) PrepareMounts() []mount.Mount {
	mounts := make([]mount.Mount, 0, len(m))

	for idx := range m {
		m := m[idx]
		containerMount := mount.Mount{
			Type:     m.Source.Type(),
			Source:   m.Source.String(),
			ReadOnly: m.ReadOnly,
			Target:   m.Target.String(),
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
