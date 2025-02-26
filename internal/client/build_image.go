package client

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
)

// buildOptions is a type that represents all options for building an image.
type buildOptions struct {
	options   types.ImageBuildOptions
	logWriter io.Writer
}

// LogWriter returns writer for build logs.
// Default: [io.Discard].
func (bo buildOptions) LogWriter() io.Writer {
	if bo.logWriter != nil {
		return bo.logWriter
	}

	return io.Discard
}

// BuildOption is a type that represents an option for building an image.
type BuildOption func(*buildOptions) error

// BuildOptions returns a build option that sets the options for building an image.
// TODO: Should we expose this or make options for each struct member?
func BuildOptions(options types.ImageBuildOptions) BuildOption {
	return func(bo *buildOptions) error {
		bo.options = options
		return nil
	}
}

// BuildLogWriter returns a build option that sets the writer for the build logs.
func BuildLogWriter(w io.Writer) BuildOption {
	return func(bo *buildOptions) error {
		bo.logWriter = w
		return nil
	}
}

// BuildImage builds an image from a build context with the specified options.
// If buildContext implements [io.Closer], it will be closed before returning.
// The first tag is returned if the build is successful.
func (c *Client) BuildImage(ctx context.Context, buildContext io.Reader, options ...BuildOption) (string, error) {
	defer func() {
		// Clean up if necessary.
		if rc, ok := buildContext.(io.Closer); ok {
			rc.Close()
		}
	}()

	if err := c.initOnce(ctx); err != nil {
		return "", fmt.Errorf("init: %w", err)
	}

	var opts buildOptions
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return "", err
		}
	}

	resp, err := backoff.RetryNotifyWithData(
		func() (*types.ImageBuildResponse, error) {
			resp, err := c.client.ImageBuild(ctx, buildContext, opts.options)
			if err != nil {
				if isPermanentClientError(err) {
					return nil, backoff.Permanent(err)
				}

				// Retryable error.
				return nil, err
			}

			return &resp, nil
		},
		backoff.WithContext(backoff.NewExponentialBackOff(), ctx),
		func(err error, _ time.Duration) {
			c.log.DebugContext(ctx, "build image", "error", err)
		},
	)
	if err != nil {
		return "", fmt.Errorf("build image: %w", err)
	}
	defer resp.Body.Close()

	// Always process the output, even if it is not printed to ensure that errors
	// during the build process are correctly handled.
	output := opts.LogWriter()
	termFd, isTerm := term.GetFdInfo(output)
	if err = jsonmessage.DisplayJSONMessagesStream(resp.Body, output, termFd, isTerm, nil); err != nil {
		return "", fmt.Errorf("build image: %w", err)
	}

	// The first tag is the one we want.
	return opts.options.Tags[0], nil
}
