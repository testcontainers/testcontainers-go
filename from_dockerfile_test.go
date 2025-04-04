package testcontainers

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/log"
)

func TestBuildImageFromDockerfile(t *testing.T) {
	provider, err := NewDockerProvider()
	require.NoError(t, err)
	defer provider.Close()

	cli := provider.Client()

	ctx := context.Background()

	tag, err := provider.BuildImage(ctx, &ContainerRequest{
		// fromDockerfileIncludingRepo {
		FromDockerfile: FromDockerfile{
			Context:    "testdata",
			Dockerfile: "echo.Dockerfile",
			Repo:       "test-repo",
			Tag:        "test-tag",
		},
		// }
	})
	require.NoError(t, err)
	require.Equal(t, "test-repo:test-tag", tag)

	_, err = cli.ImageInspect(ctx, tag)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := cli.ImageRemove(ctx, tag, image.RemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		require.NoError(t, err)
	})
}

func TestBuildImageFromDockerfile_NoRepo(t *testing.T) {
	provider, err := NewDockerProvider()
	require.NoError(t, err)
	defer provider.Close()

	cli := provider.Client()

	ctx := context.Background()

	tag, err := provider.BuildImage(ctx, &ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "testdata",
			Dockerfile: "echo.Dockerfile",
			Repo:       "test-repo",
		},
	})
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(tag, "test-repo:"))

	_, err = cli.ImageInspect(ctx, tag)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := cli.ImageRemove(ctx, tag, image.RemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		require.NoError(t, err)
	})
}

func TestBuildImageFromDockerfile_BuildError(t *testing.T) {
	ctx := context.Background()
	dockerClient, err := NewDockerClientWithOpts(ctx)
	require.NoError(t, err)

	defer dockerClient.Close()

	req := ContainerRequest{
		FromDockerfile: FromDockerfile{
			Dockerfile: "error.Dockerfile",
			Context:    filepath.Join(".", "testdata"),
		},
	}
	ctr, err := GenericContainer(ctx, GenericContainerRequest{
		ProviderType:     providerType,
		ContainerRequest: req,
		Started:          true,
	})
	CleanupContainer(t, ctr)
	require.EqualError(t, err, `create container: build image: The command '/bin/sh -c exit 1' returned a non-zero code: 1`)
}

func TestBuildImageFromDockerfile_NoTag(t *testing.T) {
	provider, err := NewDockerProvider()
	require.NoError(t, err)
	defer provider.Close()

	cli := provider.Client()

	ctx := context.Background()

	tag, err := provider.BuildImage(ctx, &ContainerRequest{
		FromDockerfile: FromDockerfile{
			Context:    "testdata",
			Dockerfile: "echo.Dockerfile",
			Tag:        "test-tag",
		},
	})
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(tag, ":test-tag"))

	_, err = cli.ImageInspect(ctx, tag)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := cli.ImageRemove(ctx, tag, image.RemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		require.NoError(t, err)
	})
}

func TestBuildImageFromDockerfile_Target(t *testing.T) {
	// there are three targets: target0, target1 and target2.
	for i := 0; i < 3; i++ {
		ctx := context.Background()
		c, err := GenericContainer(ctx, GenericContainerRequest{
			ContainerRequest: ContainerRequest{
				FromDockerfile: FromDockerfile{
					Context:    "testdata",
					Dockerfile: "target.Dockerfile",
					KeepImage:  false,
					BuildOptionsModifier: func(buildOptions *types.ImageBuildOptions) {
						buildOptions.Target = fmt.Sprintf("target%d", i)
					},
				},
			},
			Started: true,
		})
		CleanupContainer(t, c)
		require.NoError(t, err)

		r, err := c.Logs(ctx)
		require.NoError(t, err)

		logs, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("target%d\n\n", i), string(logs))
	}
}

func ExampleGenericContainer_buildFromDockerfile() {
	ctx := context.Background()

	// buildFromDockerfileWithModifier {
	c, err := Run(ctx, "",
		WithDockerfile(FromDockerfile{
			Context:    "testdata",
			Dockerfile: "target.Dockerfile",
			KeepImage:  false,
			BuildOptionsModifier: func(buildOptions *types.ImageBuildOptions) {
				buildOptions.Target = "target2"
			},
		}),
	)
	// }
	defer func() {
		if err := TerminateContainer(c); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %v", err)
		return
	}

	r, err := c.Logs(ctx)
	if err != nil {
		log.Printf("failed to get logs: %v", err)
		return
	}

	logs, err := io.ReadAll(r)
	if err != nil {
		log.Printf("failed to read logs: %v", err)
		return
	}

	fmt.Println(string(logs))

	// Output: target2
}

func TestBuildImageFromDockerfile_TargetDoesNotExist(t *testing.T) {
	// the context cancellation will happen with enough time for the build to fail.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ctr, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			FromDockerfile: FromDockerfile{
				Context:    "testdata",
				Dockerfile: "target.Dockerfile",
				KeepImage:  false,
				BuildOptionsModifier: func(buildOptions *types.ImageBuildOptions) {
					buildOptions.Target = "target-foo"
				},
			},
		},
		Started: true,
	})
	CleanupContainer(t, ctr)
	require.Error(t, err)
}
