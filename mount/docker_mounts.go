package mount

import (
	"github.com/docker/docker/api/types/mount"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

var mountTypeMapping = map[Type]mount.Type{
	TypeVolume: mount.TypeVolume,
	TypeTmpfs:  mount.TypeTmpfs,
	TypePipe:   mount.TypeNamedPipe,
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
	mounts := make([]mount.Mount, 0, len(m))

	for idx := range m {
		m := m[idx]

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
