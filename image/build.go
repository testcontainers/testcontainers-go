package image

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// BuildInfo defines what is needed to build an image
type BuildInfo interface {
	Printf(format string, args ...interface{})      // print formatted string
	BuildOptions() (types.ImageBuildOptions, error) // converts the BuildInfo to a types.ImageBuildOptions
	GetContext() (io.Reader, error)                 // the path to the build context
	GetDockerfile() string                          // the relative path to the Dockerfile, including the fileitself
	GetRepo() string                                // get repo label for image
	GetTag() string                                 // get tag label for image
	ShouldPrintBuildLog() bool                      // allow build log to be printed to stdout
	ShouldBuildImage() bool                         // return true if the image needs to be built
	GetBuildArgs() map[string]*string               // return the environment args used to build the from Dockerfile
}

// Build will build and image from context and Dockerfile, then return the tag
func Build(ctx context.Context, img BuildInfo) (string, error) {
	var buildOptions types.ImageBuildOptions

	cli, err := core.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()

	resp, err := backoff.RetryNotifyWithData(
		func() (types.ImageBuildResponse, error) {
			var err error
			buildOptions, err = img.BuildOptions()
			if err != nil {
				return types.ImageBuildResponse{}, backoff.Permanent(err)
			}
			defer tryClose(buildOptions.Context) // release resources in any case

			resp, err := cli.ImageBuild(ctx, buildOptions.Context, buildOptions)
			if err != nil {
				if core.IsPermanentClientError(err) {
					return types.ImageBuildResponse{}, backoff.Permanent(err)
				}

				img.Printf("Failed to build image: %s, will retry", err)
				return types.ImageBuildResponse{}, err
			}

			return resp, nil
		}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx),
		func(err error, duration time.Duration) {
			img.Printf("Failed to build image: %s, will retry in %s", err, duration)
		},
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if img.ShouldPrintBuildLog() {
		termFd, isTerm := term.GetFdInfo(os.Stderr)
		err = jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil)
		if err != nil {
			return "", err
		}
	}

	// need to read the response from Docker, I think otherwise the image
	// might not finish building before continuing to execute here
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return "", err
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
