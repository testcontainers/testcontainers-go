package dockerregistry

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// DockerRegistryContainer represents the DockerRegistry container type used in the module
type DockerRegistryContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the DockerRegistry container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*DockerRegistryContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "docker.io/registry:latest",
		ExposedPorts: []string{"5000/tcp"},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &DockerRegistryContainer{Container: container}, nil
}

// WithAuthentication customizer that will retrieve a htpasswd file into the htpasswd directory given in input
func WithAuthentication(htpasswdPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		req.Env["REGISTRY_AUTH"] = "htpasswd"
		req.Env["REGISTRY_AUTH_HTPASSWD_REALM"] = "Registry"
		req.Env["REGISTRY_AUTH_HTPASSWD_PATH"] = "/auth/htpasswd"

		if req.Mounts == nil {
			req.Mounts = testcontainers.ContainerMounts{}
		}

		htpasswdMount := testcontainers.ContainerMount{
			Source: testcontainers.GenericBindMountSource{
				HostPath: htpasswdPath,
			},
			Target: "/auth",
		}
		req.Mounts = append(req.Mounts, htpasswdMount)
	}

}

// WithData customizer that will retrieve a data directory and mount it inside the registry container
func WithData(dataDir string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		req.Env["REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY"] = "/data"

		if req.Mounts == nil {
			req.Mounts = testcontainers.ContainerMounts{}
		}

		dataMount := testcontainers.ContainerMount{
			Source: testcontainers.GenericBindMountSource{
				HostPath: dataDir,
			},
			Target: "/data",
		}

		req.Mounts = append(req.Mounts, dataMount)

	}
}
