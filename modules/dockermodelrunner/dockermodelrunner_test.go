package dockermodelrunner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner"
)

const (
	testModelNamespace       = "ai"
	testModelName            = "llama3.2"
	testModelTag             = "latest"
	testModelFQMN            = testModelNamespace + "/" + testModelName + ":" + testModelTag
	testModelNameNonExistent = "non-existent"
	testNonExistentFQMN      = testModelNamespace + "/" + testModelNameNonExistent + ":" + testModelTag
)

func TestDockerModelRunner(t *testing.T) {
	ctx := context.Background()

	ctr, err := dockermodelrunner.Run(ctx, "alpine/socat:1.8.0.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}

func TestRun_client(t *testing.T) {
	ctx := context.Background()

	ctr, err := dockermodelrunner.Run(ctx, "alpine/socat:1.8.0.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("model-pull", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			err := ctr.PullModel(ctx, testModelFQMN)
			require.NoError(t, err)
		})

		t.Run("failure", func(t *testing.T) {
			err := ctr.PullModel(ctx, testNonExistentFQMN)
			require.Error(t, err)
		})
	})

	t.Run("model-get", func(t *testing.T) {
		err := ctr.PullModel(ctx, testModelFQMN)
		require.NoError(t, err)

		t.Run("success", func(t *testing.T) {
			model, err := ctr.GetModel(ctx, testModelNamespace, testModelName)
			require.NoError(t, err)
			require.NotEmpty(t, model)
		})

		t.Run("failure", func(t *testing.T) {
			_, err := ctr.GetModel(ctx, testModelNamespace, testModelNameNonExistent)
			require.Error(t, err)
		})
	})

	t.Run("list-models", func(t *testing.T) {
		err := ctr.PullModel(ctx, testModelFQMN)
		require.NoError(t, err)

		t.Run("success", func(t *testing.T) {
			models, err := ctr.ListModels(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, models)

			allTags := []string{}
			for _, model := range models {
				allTags = append(allTags, model.Tags...)
			}

			require.Contains(t, allTags, testModelFQMN)
		})
	})
}
