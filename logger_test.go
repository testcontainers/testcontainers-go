package testcontainers

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestWithLogger(t *testing.T) {
	logger := TestLogger(t)
	logOpt := WithLogger(logger)
	t.Run("container", func(t *testing.T) {
		var req GenericContainerRequest
		assert.NilError(t, logOpt.Customize(&req))
		assert.DeepEqual(t, logger, req.Logger)
	})

	t.Run("provider", func(t *testing.T) {
		var opts GenericProviderOptions
		logOpt.ApplyGenericTo(&opts)
		assert.DeepEqual(t, logger, opts.Logger)
	})

	t.Run("docker", func(t *testing.T) {
		opts := &DockerProviderOptions{
			GenericProviderOptions: &GenericProviderOptions{},
		}
		logOpt.ApplyDockerTo(opts)
		assert.DeepEqual(t, logger, opts.Logger)
	})
}
