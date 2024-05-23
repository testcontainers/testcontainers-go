package testcontainers

import (
	"github.com/docker/docker/api/types/mount"
	tcmount "github.com/testcontainers/testcontainers-go/mount"
)

// Deprecated: will be removed in a future release
var mountTypeMapping = map[tcmount.Type]mount.Type{
	tcmount.TypeBind:   mount.TypeBind, // Deprecated, it will be removed in a future release
	tcmount.TypeVolume: mount.TypeVolume,
	tcmount.TypeTmpfs:  mount.TypeTmpfs,
	tcmount.TypePipe:   mount.TypeNamedPipe,
}

// Deprecated: use tcmount.BindMounter instead
// BindMounter can optionally be implemented by mount sources
// to support advanced scenarios based on mount.BindOptions
type BindMounter = tcmount.BindMounter

// Deprecated: use tcmount.VolumeMounter instead
// VolumeMounter can optionally be implemented by mount sources
// to support advanced scenarios based on mount.VolumeOptions
type VolumeMounter = tcmount.VolumeMounter

// Deprecated: use tcmount.TmpfsMounter instead
// TmpfsMounter can optionally be implemented by mount sources
// to support advanced scenarios based on mount.TmpfsOptions
type TmpfsMounter = tcmount.TmpfsMounter

// Deprecated: use tcmount.DockerBindSource instead
type DockerBindMountSource = tcmount.DockerBindSource

// Deprecated: use tcmount.DockerVolumeSource instead
type DockerVolumeMountSource = tcmount.DockerVolumeSource

// Deprecated: use tcmount.DockerTmpfsSource instead
type DockerTmpfsMountSource = tcmount.DockerTmpfsSource
