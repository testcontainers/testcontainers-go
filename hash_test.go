package testcontainers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

func TestHashContainerRequest(t *testing.T) {
	req := ContainerRequest{
		Image: "nginx",
		Env:   map[string]string{"a": "b"},
		FromDockerfile: FromDockerfile{
			BuildOptionsModifier: func(options *types.ImageBuildOptions) {},
		},
		ExposedPorts:      []string{"80/tcp"},
		Privileged:        false,
		ImageSubstitutors: []ImageSubstitutor{newPrependHubRegistry("localhost:5000")},
		LifecycleHooks: []ContainerLifecycleHooks{
			{
				PreStarts: []ContainerHook{
					func(ctx context.Context, c Container) error {
						return nil
					},
				},
			},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			// NOOP
		},
		WaitingFor: wait.ForLog("nginx: ready"),
		Files: []ContainerFile{
			{
				HostFilePath:      filepath.Join("testdata", "hello.sh"),
				ContainerFilePath: "/hello.sh",
				FileMode:          0o755,
			},
		},
	}

	hash1 := req.hash()
	require.NotEqual(t, 0, hash1)

	hash2 := req.hash()
	require.NotEqual(t, 0, hash2)

	require.Equal(t, hash1.Hash, hash2.Hash)
	require.Equal(t, hash1.FilesHash, hash2.FilesHash)
}

func TestHashContainerRequest_includingDirs(t *testing.T) {
	req1 := ContainerRequest{
		Image: "nginx",
		Env:   map[string]string{"a": "b"},
		FromDockerfile: FromDockerfile{
			BuildOptionsModifier: func(options *types.ImageBuildOptions) {},
		},
		ExposedPorts:      []string{"80/tcp"},
		Privileged:        false,
		ImageSubstitutors: []ImageSubstitutor{newPrependHubRegistry("localhost:5000")},
		LifecycleHooks: []ContainerLifecycleHooks{
			{
				PreStarts: []ContainerHook{
					func(ctx context.Context, c Container) error {
						return nil
					},
				},
			},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			// NOOP
		},
		WaitingFor: wait.ForLog("nginx: ready"),
		Files: []ContainerFile{
			{
				HostFilePath:      "testdata",
				ContainerFilePath: "/data",
				FileMode:          0o755,
			},
		},
	}

	req2 := ContainerRequest{
		Image: "nginx",
		Env:   map[string]string{"a": "b"},
		FromDockerfile: FromDockerfile{
			BuildOptionsModifier: func(options *types.ImageBuildOptions) {},
		},
		ExposedPorts:      []string{"80/tcp"},
		Privileged:        false,
		ImageSubstitutors: []ImageSubstitutor{newPrependHubRegistry("localhost:5000")},
		LifecycleHooks: []ContainerLifecycleHooks{
			{
				PreStarts: []ContainerHook{
					func(ctx context.Context, c Container) error {
						return nil
					},
				},
			},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			// NOOP
		},
		WaitingFor: wait.ForLog("nginx: ready"),
		Files: []ContainerFile{
			{
				HostFilePath:      filepath.Join("testdata", "data"),
				ContainerFilePath: "/data",
				FileMode:          0o755,
			},
		},
	}

	hash1 := req1.hash()
	require.NotEqual(t, 0, hash1)

	hash2 := req2.hash()
	require.NotEqual(t, 0, hash2)

	require.NotEqual(t, hash1.Hash, hash2.Hash) // because the hostfile path is different
	require.Zero(t, hash1.FilesHash)            // the dir is not included in the hash
	require.Equal(t, hash1.FilesHash, hash2.FilesHash)
}

func TestHashContainerRequest_differs(t *testing.T) {
	req1 := ContainerRequest{
		Image: "nginx",
		Env:   map[string]string{"a": "b"},
		FromDockerfile: FromDockerfile{
			BuildOptionsModifier: func(options *types.ImageBuildOptions) {},
		},
		ExposedPorts:      []string{"80/tcp"},
		Privileged:        false,
		ImageSubstitutors: []ImageSubstitutor{newPrependHubRegistry("localhost:5000")},
		LifecycleHooks: []ContainerLifecycleHooks{
			{
				PreStarts: []ContainerHook{
					func(ctx context.Context, c Container) error {
						return nil
					},
				},
			},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			// NOOP
		},
		WaitingFor: wait.ForLog("nginx: ready"),
		Files: []ContainerFile{
			{
				HostFilePath:      filepath.Join("testdata", "hello.sh"),
				ContainerFilePath: "/hello.sh",
				FileMode:          0o755,
			},
		},
	}

	req2 := ContainerRequest{
		Image: "nginx1", // this is the only difference with req1
		Env:   map[string]string{"a": "b"},
		FromDockerfile: FromDockerfile{
			BuildOptionsModifier: func(options *types.ImageBuildOptions) {},
		},
		ExposedPorts:      []string{"80/tcp"},
		Privileged:        false,
		ImageSubstitutors: []ImageSubstitutor{newPrependHubRegistry("localhost:5000")},
		LifecycleHooks: []ContainerLifecycleHooks{
			{
				PreStarts: []ContainerHook{
					func(ctx context.Context, c Container) error {
						return nil
					},
				},
			},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			// NOOP
		},
		WaitingFor: wait.ForLog("nginx: ready"),
		Files: []ContainerFile{
			{
				HostFilePath:      filepath.Join("testdata", "hello.sh"),
				ContainerFilePath: "/hello.sh",
				FileMode:          0o755,
			},
		},
	}

	hash1 := req1.hash()
	require.NotEqual(t, 0, hash1)

	hash2 := req2.hash()
	require.NotEqual(t, 0, hash2)

	require.NotEqual(t, hash1.Hash, hash2.Hash)
	require.Equal(t, hash1.FilesHash, hash2.FilesHash)
}

func TestHashContainerRequest_modifiedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// create a temporary file to be used in the test
	tmpFile := filepath.Join(tmpDir, "hello.sh")

	f, err := os.Create(tmpFile)
	require.NoError(t, err)

	_, err = f.WriteString("echo hello gopher!")
	require.NoError(t, err)

	req := ContainerRequest{
		Image: "nginx",
		Env:   map[string]string{"a": "b"},
		FromDockerfile: FromDockerfile{
			BuildOptionsModifier: func(options *types.ImageBuildOptions) {},
		},
		ExposedPorts:      []string{"80/tcp"},
		Privileged:        false,
		ImageSubstitutors: []ImageSubstitutor{newPrependHubRegistry("localhost:5000")},
		LifecycleHooks: []ContainerLifecycleHooks{
			{
				PreStarts: []ContainerHook{
					func(ctx context.Context, c Container) error {
						return nil
					},
				},
			},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			// NOOP
		},
		WaitingFor: wait.ForLog("nginx: ready"),
		Files: []ContainerFile{
			{
				HostFilePath:      tmpFile,
				ContainerFilePath: "/hello.sh",
				FileMode:          0o755,
			},
		},
	}

	hash1 := req.hash()
	require.NotEqual(t, 0, hash1)

	// modify the initial file
	_, err = f.WriteString("echo goodbye gopher!")
	require.NoError(t, err)

	hash2 := req.hash()
	require.NotEqual(t, 0, hash2)

	require.Equal(t, hash1.Hash, hash2.Hash)
	require.NotEqual(t, hash1.FilesHash, hash2.FilesHash)
}
