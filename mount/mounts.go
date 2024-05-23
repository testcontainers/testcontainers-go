package mount

import "errors"

const (
	TypeBind Type = iota // Deprecated: Use TypeVolume instead
	TypeVolume
	TypeTmpfs
	TypePipe
)

var (
	ErrDuplicateMountTarget = errors.New("duplicate mount target detected")
	ErrInvalidBindMount     = errors.New("invalid bind mount")
)

var (
	_ ContainerSource = (*GenericBindSource)(nil) // Deprecated: use Files or HostConfigModifier in the ContainerRequest, or copy files container APIs to make containers portable across Docker environments
	_ ContainerSource = (*GenericVolumeSource)(nil)
	_ ContainerSource = (*GenericTmpfsSource)(nil)
)

type (
	// ContainerMounts represents a collection of mounts for a container
	ContainerMounts []ContainerMount
	Type            uint
)

// ContainerSource is the base for all mount sources
type ContainerSource interface {
	// Source will be used as Source field in the final mount
	// this might either be a volume name, a host path or might be empty e.g. for Tmpfs
	Source() string

	// Type determines the final mount type
	// possible options are limited by the Docker API
	Type() Type
}

// Deprecated: use Files or HostConfigModifier in the ContainerRequest, or copy files container APIs to make containers portable across Docker environments
// GenericBindSource implements ContainerSource and represents a bind mount
// Optionally mount.BindOptions might be added for advanced scenarios
type GenericBindSource struct {
	// HostPath is the path mounted into the container
	// the same host path might be mounted to multiple locations within a single container
	HostPath string
}

// Deprecated: use Files or HostConfigModifier in the ContainerRequest, or copy files container APIs to make containers portable across Docker environments
func (s GenericBindSource) Source() string {
	return s.HostPath
}

// Deprecated: use Files or HostConfigModifier in the ContainerRequest, or copy files container APIs to make containers portable across Docker environments
func (GenericBindSource) Type() Type {
	return TypeBind
}

// GenericVolumeSource implements ContainerSource and represents a volume mount
type GenericVolumeSource struct {
	// Name refers to the name of the volume to be mounted
	// the same volume might be mounted to multiple locations within a single container
	Name string
}

func (s GenericVolumeSource) Source() string {
	return s.Name
}

func (GenericVolumeSource) Type() Type {
	return TypeVolume
}

// GenericTmpfsSource implements ContainerSource and represents a TmpFS mount
// Optionally mount.TmpfsOptions might be added for advanced scenarios
type GenericTmpfsSource struct{}

func (s GenericTmpfsSource) Source() string {
	return ""
}

func (GenericTmpfsSource) Type() Type {
	return TypeTmpfs
}

// ContainerTarget represents the target path within a container where the mount will be available
// Note that mount targets must be unique. It's not supported to mount different sources to the same target.
type ContainerTarget string

func (t ContainerTarget) Target() string {
	return string(t)
}

// Deprecated: use Files or HostConfigModifier in the ContainerRequest, or copy files container APIs to make containers portable across Docker environments
// BindMount returns a new ContainerMount with a GenericBindMountSource as source
// This is a convenience method to cover typical use cases.
func BindMount(hostPath string, mountTarget ContainerTarget) ContainerMount {
	return ContainerMount{
		Source: GenericBindSource{HostPath: hostPath},
		Target: mountTarget,
	}
}

// VolumeMount returns a new ContainerMount with a GenericVolumeMountSource as source
// This is a convenience method to cover typical use cases.
func VolumeMount(volumeName string, mountTarget ContainerTarget) ContainerMount {
	return ContainerMount{
		Source: GenericVolumeSource{Name: volumeName},
		Target: mountTarget,
	}
}

// Mounts returns a ContainerMounts to support a more fluent API
func Mounts(mounts ...ContainerMount) ContainerMounts {
	return mounts
}

// ContainerMount models a mount into a container
type ContainerMount struct {
	// Source is typically either a GenericVolumeSource, as BindMount is not supported by all Docker environments
	Source ContainerSource
	// Target is the path where the mount should be mounted within the container
	Target ContainerTarget
	// ReadOnly determines if the mount should be read-only
	ReadOnly bool
}
