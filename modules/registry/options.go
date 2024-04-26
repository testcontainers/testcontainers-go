package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
)

const (
	containerDataPath     string = "/data"
	containerHtpasswdPath string = "/auth/htpasswd"
)

// WithData is a custom option to set the data directory for the registry,
// which is used to store the images. It will copy the data from the host to
// the container in the /data path. The container will be configured to use
// this path as the root directory for the registry, thanks to the
// REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY environment variable.
// The dataPath must have the same structure as the registry data directory.
func WithData(dataPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      dataPath,
			ContainerFilePath: containerDataPath,
		})

		req.Env["REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY"] = containerDataPath
		return nil
	}
}

// WithHtpasswd is a custom option to set the htpasswd credentials for the registry
// It will create a temporary file with the credentials and copy it to the container
// in the /auth/htpasswd path. The container will be configured to use this file as
// the htpasswd file, thanks to the REGISTRY_AUTH_HTPASSWD_PATH environment variable.
func WithHtpasswd(credentials string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		tmpFile, err := os.Create(filepath.Join(os.TempDir(), "htpasswd"))
		if err != nil {
			tmpFile, err = os.Create(".")
			if err != nil {
				return fmt.Errorf("cannot create the file in the temp dir or in the current dir: %w", err)
			}
		}
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(credentials)
		if err != nil {
			return fmt.Errorf("cannot write the credentials to the file: %w", err)
		}

		return WithHtpasswdFile(tmpFile.Name())(req)
	}
}

// WithHtpasswdFile is a custom option to set the htpasswd file for the registry
// It will copy a file with the credentials in the /auth/htpasswd path.
// The container will be configured to use this file as the htpasswd file,
// thanks to the REGISTRY_AUTH_HTPASSWD_PATH environment variable.
func WithHtpasswdFile(htpasswdPath string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      htpasswdPath,
			ContainerFilePath: containerHtpasswdPath,
			FileMode:          0o644,
		})

		req.Env["REGISTRY_AUTH"] = "htpasswd"
		req.Env["REGISTRY_AUTH_HTPASSWD_REALM"] = "Registry"
		req.Env["REGISTRY_AUTH_HTPASSWD_PATH"] = containerHtpasswdPath
		req.Env["REGISTRY_AUTH_HTPASSWD_PATH"] = containerHtpasswdPath
		return nil
	}
}
