package image

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/docker/api/types"

	"github.com/testcontainers/testcontainers-go/auth"
	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/log"
)

// Pull tries to pull the image while respecting the ctx cancellations.
// Besides, if the image cannot be pulled due to ErrorNotFound then no need to retry but terminate immediately.
func Pull(ctx context.Context, tag string, logger log.Logging, pullOpt types.ImagePullOptions) error {
	registry, imageAuth, err := auth.ForDockerImage(ctx, tag)
	if err != nil {
		logger.Printf("Failed to get image auth for %s. Setting empty credentials for the image: %s. Error is:%s", registry, tag, err)
	} else {
		// see https://github.com/docker/docs/blob/e8e1204f914767128814dca0ea008644709c117f/engine/api/sdk/examples.md?plain=1#L649-L657
		encodedJSON, err := json.Marshal(imageAuth)
		if err != nil {
			logger.Printf("Failed to marshal image auth. Setting empty credentials for the image: %s. Error is:%s", tag, err)
		} else {
			pullOpt.RegistryAuth = base64.URLEncoding.EncodeToString(encodedJSON)
		}
	}

	cli, err := core.NewClient(ctx)
	if err != nil {
		return err
	}
	defer cli.Close()

	var pull io.ReadCloser
	err = backoff.Retry(func() error {
		pull, err = cli.ImagePull(ctx, tag, pullOpt)
		if err != nil {
			if core.IsPermanentClientError(err) {
				return backoff.Permanent(err)
			}
			logger.Printf("Failed to pull image: %s, will retry", err)
			return err
		}

		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
	if err != nil {
		return err
	}
	defer pull.Close()

	// download of docker image finishes at EOF of the pull request
	_, err = io.ReadAll(pull)
	return err
}
