package image

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"

	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/log"
)

// Builder defines what is needed to build an image
type Builder interface {
	BuildOptions() (types.ImageBuildOptions, error) // converts the Builder to a types.ImageBuildOptions
	GetContext() (io.Reader, error)                 // the path to the build context
	GetDockerfile() string                          // the relative path to the Dockerfile, including the file itself
	GetRepo() string                                // get repo label for image
	GetTag() string                                 // get tag label for image
	BuildLogWriter() io.Writer                      // for output of build log, use io.Discard to disable the output
	ShouldBuildImage() bool                         // return true if the image needs to be built
	GetBuildArgs() map[string]*string               // return the environment args used to build the Dockerfile
	GetAuthConfigs() map[string]registry.AuthConfig // Deprecated. Testcontainers will detect registry credentials automatically. Return the auth configs to be able to pull from an authenticated docker registry
}

// Build build an image from the given Builder, using the default docker client.
// It returns the first tag of the image.
func Build(ctx context.Context, img Builder) (string, error) {
	cli, err := core.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("new client: %w", err)
	}
	defer cli.Close()

	return buildWithClient(ctx, cli, img)
}

// buildWithClient build and image from context and Dockerfile, then return the first tag of the image.
// The caller is responsible for closing the docker client.
func buildWithClient(ctx context.Context, dockerClient client.APIClient, img Builder) (string, error) {
	var buildOptions types.ImageBuildOptions
	resp, err := backoff.RetryNotifyWithData(
		func() (types.ImageBuildResponse, error) {
			var err error
			buildOptions, err = img.BuildOptions()
			if err != nil {
				return types.ImageBuildResponse{}, backoff.Permanent(fmt.Errorf("build options: %w", err))
			}
			defer tryClose(buildOptions.Context) // release resources in any case

			resp, err := dockerClient.ImageBuild(ctx, buildOptions.Context, buildOptions)
			if err != nil {
				if core.IsPermanentClientError(err) {
					return types.ImageBuildResponse{}, backoff.Permanent(fmt.Errorf("build image: %w", err))
				}
				return types.ImageBuildResponse{}, err
			}

			return resp, nil
		},
		backoff.WithContext(backoff.NewExponentialBackOff(), ctx),
		func(err error, _ time.Duration) {
			log.Printf("Failed to build image: %s, will retry", err)
		},
	)
	if err != nil {
		return "", err // Error is already wrapped.
	}
	defer resp.Body.Close()

	output := img.BuildLogWriter()

	// Always process the output, even if it is not printed
	// to ensure that errors during the build process are
	// correctly handled.
	termFd, isTerm := term.GetFdInfo(output)
	if err = jsonmessage.DisplayJSONMessagesStream(resp.Body, output, termFd, isTerm, nil); err != nil {
		return "", fmt.Errorf("build image: %w", err)
	}

	// the first tag is the one we want
	return buildOptions.Tags[0], nil
}

func tryClose(r io.Reader) {
	rc, ok := r.(io.Closer)
	if ok {
		_ = rc.Close()
	}
}
