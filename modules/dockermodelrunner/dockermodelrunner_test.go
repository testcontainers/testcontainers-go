package dockermodelrunner_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner"
	"github.com/testcontainers/testcontainers-go/modules/socat"
)

const (
	testModelNamespace       = "ai"
	testModelName            = "smollm2"
	testModelTag             = "360M-Q4_K_M"
	testModelFQMN            = testModelNamespace + "/" + testModelName + ":" + testModelTag
	testModelNameNonExistent = "non-existent"
	testNonExistentFQMN      = testModelNamespace + "/" + testModelNameNonExistent + ":" + testModelTag
)

func TestRun(t *testing.T) {
	skipIfDockerDesktopNotRunning(t)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		ctr, err := dockermodelrunner.Run(ctx)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)
	})

	t.Run("success/with-image", func(t *testing.T) {
		ctr, err := dockermodelrunner.Run(ctx, testcontainers.WithImage(socat.DefaultImage))
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)
	})

	t.Run("failure/with-image", func(t *testing.T) {
		ctr, err := dockermodelrunner.Run(ctx, testcontainers.WithImage("alpine:latest"))
		testcontainers.CleanupContainer(t, ctr)
		require.Error(t, err)
	})
}

func TestRun_client(t *testing.T) {
	skipIfDockerDesktopNotRunning(t)
	ctx := context.Background()

	ctr, err := dockermodelrunner.Run(ctx)
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

		t.Run("failure/timeout", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
			defer cancel()
			err := ctr.PullModel(ctx, testModelFQMN)
			require.Error(t, err)
		})
	})

	t.Run("model-inspect", func(t *testing.T) {
		err := ctr.PullModel(ctx, testModelFQMN)
		require.NoError(t, err)

		t.Run("success", func(t *testing.T) {
			model, err := ctr.InspectModel(ctx, testModelNamespace, testModelName+":"+testModelTag)
			require.NoError(t, err)
			require.NotNil(t, model)
		})

		t.Run("failure", func(t *testing.T) {
			model, err := ctr.InspectModel(ctx, testModelNamespace, testModelNameNonExistent)
			require.Error(t, err)
			require.Nil(t, model)
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
