package testcontainers

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildImageFromDockerfile(t *testing.T) {
	provider, err := NewDockerProvider()
	if err != nil {
		t.Fatal(err)
	}
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
	assert.Equal(t, "test-repo:test-tag", tag)

	_, _, err = cli.ImageInspectWithRaw(ctx, tag)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := cli.ImageRemove(ctx, tag, types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestBuildImageFromDockerfile_NoRepo(t *testing.T) {
	provider, err := NewDockerProvider()
	if err != nil {
		t.Fatal(err)
	}
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
	assert.True(t, strings.HasPrefix(tag, "test-repo:"))

	_, _, err = cli.ImageInspectWithRaw(ctx, tag)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := cli.ImageRemove(ctx, tag, types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestBuildImageFromDockerfile_NoTag(t *testing.T) {
	provider, err := NewDockerProvider()
	if err != nil {
		t.Fatal(err)
	}
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
	assert.True(t, strings.HasSuffix(tag, ":test-tag"))

	_, _, err = cli.ImageInspectWithRaw(ctx, tag)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := cli.ImageRemove(ctx, tag, types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestBuildImageFromDockerfile_Target(t *testing.T) {
	// there are thre targets: target0, target1 and target2.
	for i := 0; i < 3; i++ {
		ctx := context.Background()
		c, err := GenericContainer(ctx, GenericContainerRequest{
			ContainerRequest: ContainerRequest{
				FromDockerfile: FromDockerfile{
					Context:       "testdata",
					Dockerfile:    "target.Dockerfile",
					PrintBuildLog: true,
					KeepImage:     false,
					BuildOptionsModifier: func(buildOptions *types.ImageBuildOptions) {
						buildOptions.Target = fmt.Sprintf("target%d", i)
					},
				},
			},
			Started: true,
		})
		require.NoError(t, err)

		r, err := c.Logs(ctx)
		require.NoError(t, err)

		logs, err := io.ReadAll(r)
		require.NoError(t, err)

		assert.Equal(t, fmt.Sprintf("target%d\n\n", i), string(logs))

		t.Cleanup(func() {
			require.NoError(t, c.Terminate(ctx))
		})
	}
}

func ExampleGenericContainer_buildFromDockerfile() {
	ctx := context.Background()

	// buildFromDockerfileWithModifier {
	c, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			FromDockerfile: FromDockerfile{
				Context:       "testdata",
				Dockerfile:    "target.Dockerfile",
				PrintBuildLog: true,
				KeepImage:     false,
				BuildOptionsModifier: func(buildOptions *types.ImageBuildOptions) {
					buildOptions.Target = "target2"
				},
			},
		},
		Started: true,
	})
	// }
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}

	r, err := c.Logs(ctx)
	if err != nil {
		log.Fatalf("failed to get logs: %v", err)
	}

	logs, err := io.ReadAll(r)
	if err != nil {
		log.Fatalf("failed to read logs: %v", err)
	}

	fmt.Println(string(logs))

	// Output: target2
}

func TestBuildImageFromDockerfile_TargetDoesNotExist(t *testing.T) {
	// the context cancellation will happen with enough time for the build to fail.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			FromDockerfile: FromDockerfile{
				Context:       "testdata",
				Dockerfile:    "target.Dockerfile",
				PrintBuildLog: true,
				KeepImage:     false,
				BuildOptionsModifier: func(buildOptions *types.ImageBuildOptions) {
					buildOptions.Target = "target-foo"
				},
			},
		},
		Started: true,
	})
	require.Error(t, err)
}
