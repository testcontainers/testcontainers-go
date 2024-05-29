package image

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"

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
	buildOptions, err := img.BuildOptions()

	cli, err := core.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()

	var buildError error
	var resp types.ImageBuildResponse
	err = backoff.Retry(func() error {
		resp, err = cli.ImageBuild(ctx, buildOptions.Context, buildOptions)
		if err != nil {
			buildError = errors.Join(buildError, err)
			if core.IsPermanentClientError(err) {
				return backoff.Permanent(err)
			}

			img.Printf("Failed to build image: %s, will retry", err)
			return err
		}

		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
	if err != nil {
		return "", errors.Join(buildError, err)
	}

	if img.ShouldPrintBuildLog() {
		termFd, isTerm := term.GetFdInfo(os.Stderr)
		err = jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil)
		if err != nil {
			return "", err
		}
	}

	// need to read the response from Docker, I think otherwise the image
	// might not finish building before continuing to execute here
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	_ = resp.Body.Close()

	// the first tag is the one we want
	return buildOptions.Tags[0], nil
}
