package firebase

import (
	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	emulators map[string]string
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Firebase container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithRoot sets the directory which is copied to the destination container as firebase root
func WithRoot(rootPath string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithFiles(testcontainers.ContainerFile{
		HostFilePath:      rootPath,
		ContainerFilePath: rootFilePath,
		FileMode:          0o775,
	})
}

// WithData names the data directory (by default under firebase root), can be used as a way of setting up fixtures.
// Usage of absolute path will imply that the user knows how to mount external directory into the container.
func WithData(dataPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["DATA_DIRECTORY"] = dataPath
		return nil
	}
}

const cacheFilePath = "/root/.cache/firebase"

// WithCache enables firebase binary cache based on session (meaningful only when multiple tests are used)
func WithCache() testcontainers.CustomizeRequestOption {
	volumeName := "firestore-cache-" + testcontainers.SessionID()

	return testcontainers.WithMounts(testcontainers.ContainerMount{
		Source: testcontainers.DockerVolumeMountSource{
			Name: volumeName,
		},
		Target: cacheFilePath,
	})
}
