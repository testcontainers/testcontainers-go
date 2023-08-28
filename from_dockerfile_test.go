package testcontainers

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
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
			Context:    filepath.Join("testdata"),
			Dockerfile: "echo.Dockerfile",
			Repo:       "test-repo",
			Tag:        "test-tag",
		},
		// }
	})
	assert.Nil(t, err)
	assert.Equal(t, "test-repo:test-tag", tag)

	_, _, err = cli.ImageInspectWithRaw(ctx, tag)
	assert.Nil(t, err)

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
			Context:    filepath.Join("testdata"),
			Dockerfile: "echo.Dockerfile",
			Repo:       "test-repo",
		},
	})
	assert.Nil(t, err)
	assert.True(t, strings.HasPrefix(tag, "test-repo:"))

	_, _, err = cli.ImageInspectWithRaw(ctx, tag)
	assert.Nil(t, err)

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
			Context:    filepath.Join("testdata"),
			Dockerfile: "echo.Dockerfile",
			Tag:        "test-tag",
		},
	})
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(tag, ":test-tag"))

	_, _, err = cli.ImageInspectWithRaw(ctx, tag)
	assert.Nil(t, err)

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
