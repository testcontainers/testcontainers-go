package mount

import (
	"github.com/docker/docker/api/types/mount"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

var mountTypeMapping = map[Type]mount.Type{
	TypeBind:   mount.TypeBind, // Deprecated, it will be removed in a future release
	TypeVolume: mount.TypeVolume,
	TypeTmpfs:  mount.TypeTmpfs,
	TypePipe:   mount.TypeNamedPipe,
}

// Deprecated: use Files or HostConfigModifier in the ContainerRequest, or copy files container APIs to make containers portable across Docker environments
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

// Deprecated: use Files or HostConfigModifier in the ContainerRequest, or copy files container APIs to make containers portable across Docker environments
type DockerBindSource struct {
	*mount.BindOptions

	// HostPath is the path mounted into the container
	// the same host path might be mounted to multiple locations within a single container
	HostPath string
}

// Deprecated: use Files or HostConfigModifier in the ContainerRequest, or copy files container APIs to make containers portable across Docker environments
func (s DockerBindSource) Source() string {
	return s.HostPath
}

// Deprecated: use Files or HostConfigModifier in the ContainerRequest, or copy files container APIs to make containers portable across Docker environments
func (DockerBindSource) Type() Type {
	return TypeBind
}

// Deprecated: use Files or HostConfigModifier in the ContainerRequest, or copy files container APIs to make containers portable across Docker environments
func (s DockerBindSource) GetBindOptions() *mount.BindOptions {
	return s.BindOptions
}

type DockerVolumeSource struct {
	*mount.VolumeOptions

	// Name refers to the name of the volume to be mounted
	// the same volume might be mounted to multiple locations within a single container
	Name string
}

func (s DockerVolumeSource) Source() string {
	return s.Name
}

func (DockerVolumeSource) Type() Type {
	return TypeVolume
}

func (s DockerVolumeSource) GetVolumeOptions() *mount.VolumeOptions {
	return s.VolumeOptions
}

type DockerTmpfsSource struct {
	GenericTmpfsSource
	*mount.TmpfsOptions
}

func (s DockerTmpfsSource) GetTmpfsOptions() *mount.TmpfsOptions {
	return s.TmpfsOptions
}

// Prepare maps the given []ContainerMount to the corresponding
// []mount.Mount for further processing
func (m ContainerMounts) Prepare() []mount.Mount {
	return mapToDockerMounts(m)
}

// mapToDockerMounts maps the given []ContainerMount to the corresponding
// []mount.Mount for further processing
func mapToDockerMounts(containerMounts ContainerMounts) []mount.Mount {
	mounts := make([]mount.Mount, 0, len(containerMounts))

	for idx := range containerMounts {
		m := containerMounts[idx]

		var mountType mount.Type
		if mt, ok := mountTypeMapping[m.Source.Type()]; ok {
			mountType = mt
		} else {
			continue
		}

		containerMount := mount.Mount{
			Type:     mountType,
			Source:   m.Source.Source(),
			ReadOnly: m.ReadOnly,
			Target:   m.Target.Target(),
		}

		switch typedMounter := m.Source.(type) {
		case VolumeMounter:
			containerMount.VolumeOptions = typedMounter.GetVolumeOptions()
		case TmpfsMounter:
			containerMount.TmpfsOptions = typedMounter.GetTmpfsOptions()
		case BindMounter:
			// Logger.Printf("Mount type %s is not supported by Testcontainers for Go", m.Source.Type())
		default:
			// The provided source type has no custom options
		}

		if mountType == mount.TypeVolume {
			if containerMount.VolumeOptions == nil {
				containerMount.VolumeOptions = &mount.VolumeOptions{
					Labels: make(map[string]string),
				}
			}
			for k, v := range core.DefaultLabels(core.SessionID()) {
				containerMount.VolumeOptions.Labels[k] = v
			}
		}

		mounts = append(mounts, containerMount)
	}

	return mounts
}
