package testcontainers

import (
	"errors"

	tcmount "github.com/testcontainers/testcontainers-go/mount"
)

const (
	MountTypeBind   MountType = iota // Deprecated: Use mount.TypeVolume instead
	MountTypeVolume                  // Deprecated: use mount.TypeVolume instead
	MountTypeTmpfs                   // Deprecated: use mount.TypeTmpfs instead
	MountTypePipe                    // Deprecated: use mount.TypeNamedPipe instead
)

var (
	// Deprecated: use tcmount.ErrDuplicateMountTarget instead
	ErrDuplicateMountTarget = errors.New("duplicate mount target detected")
	// Deprecated: use tcmount.ErrInvalidBindMount instead
	ErrInvalidBindMount = errors.New("invalid bind mount")
)

var (
	_ ContainerMountSource = (*GenericBindMountSource)(nil)   // Deprecated: will be removed in a future release
	_ ContainerMountSource = (*GenericVolumeMountSource)(nil) // Deprecated: will be removed in a future release
	_ ContainerMountSource = (*GenericTmpfsMountSource)(nil)  // Deprecated: will be removed in a future release
)

type (
	// Deprecated: use tcmount.ContainerMounts instead
	// ContainerMounts represents a collection of mounts for a container
	ContainerMounts = tcmount.ContainerMounts
	// Deprecated: use tcmount.Type instead
	MountType = tcmount.Type
)

// Deprecated: use tcmount.ContainerMountSource instead
// ContainerMountSource is the base for all mount sources
type ContainerMountSource = tcmount.ContainerSource

// Deprecated: use tcmount.GenericBindMountSource instead
// GenericBindMountSource implements ContainerMountSource and represents a bind mount
// Optionally mount.BindOptions might be added for advanced scenarios
type GenericBindMountSource = tcmount.GenericBindSource

// Deprecated: use tcmount.GenericVolumeMountSource instead
// GenericVolumeMountSource implements ContainerMountSource and represents a volume mount
type GenericVolumeMountSource = tcmount.GenericVolumeSource

// Deprecated: use tcmount.GenericTmpfsMountSource instead
// GenericTmpfsMountSource implements ContainerMountSource and represents a TmpFS mount
// Optionally mount.TmpfsOptions might be added for advanced scenarios
type GenericTmpfsMountSource = tcmount.GenericTmpfsSource

// Deprecated: use tcmount.ContainerMountTarget instead
// ContainerMountTarget represents the target path within a container where the mount will be available
// Note that mount targets must be unique. It's not supported to mount different sources to the same target.
type ContainerMountTarget = tcmount.ContainerTarget

// Deprecated: use tcmount.BindMount instead
// BindMount returns a new ContainerMount with a GenericBindMountSource as source
// This is a convenience method to cover typical use cases.
func BindMount(hostPath string, mountTarget ContainerMountTarget) ContainerMount {
	return ContainerMount{
		Source: GenericBindMountSource{HostPath: hostPath},
		Target: mountTarget,
	}
}

// Deprecated: use tcmount.VolumeMount instead
// VolumeMount returns a new ContainerMount with a GenericVolumeMountSource as source
// This is a convenience method to cover typical use cases.
func VolumeMount(volumeName string, mountTarget ContainerMountTarget) ContainerMount {
	return ContainerMount{
		Source: GenericVolumeMountSource{Name: volumeName},
		Target: mountTarget,
	}
}

// Deprecated: use tcmount.Mounts instead
// Mounts returns a ContainerMounts to support a more fluent API
func Mounts(mounts ...ContainerMount) ContainerMounts {
	return mounts
}

// Deprecated: use tcmount.ContainerMount instead
// ContainerMount models a mount into a container
type ContainerMount = tcmount.ContainerMount
