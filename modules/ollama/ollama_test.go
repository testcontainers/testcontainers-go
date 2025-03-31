package ollama_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/ollama"
)

func TestOllama(t *testing.T) {
	ctx := context.Background()

	ctr, err := ollama.Run(ctx, "ollama/ollama:0.5.7")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("ConnectionString", func(t *testing.T) {
		// connectionString {
		connectionStr, err := ctr.ConnectionString(ctx)
		// }
		require.NoError(t, err)

		httpClient := &http.Client{}
		resp, err := httpClient.Get(connectionStr)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Pull and Run Model", func(t *testing.T) {
		model := "all-minilm"

		_, _, err = ctr.Exec(context.Background(), []string{"ollama", "pull", model})
		require.NoError(t, err)

		_, _, err = ctr.Exec(context.Background(), []string{"ollama", "run", model})
		require.NoError(t, err)

		assertLoadedModel(t, ctr)
	})

	t.Run("Commit to image including model", func(t *testing.T) {
		// commitOllamaContainer {

		// Defining the target image name based on the default image and a random string.
		// Users can change the way this is generated, but it should be unique.
		targetImage := fmt.Sprintf("%s-%s", ollama.DefaultOllamaImage, strings.ToLower(uuid.New().String()[:4]))

		err := ctr.Commit(context.Background(), targetImage)
		// }
		require.NoError(t, err)

		newOllamaContainer, err := ollama.Run(
			context.Background(),
			targetImage,
		)
		testcontainers.CleanupContainer(t, newOllamaContainer)
		require.NoError(t, err)

		assertLoadedModel(t, newOllamaContainer)
	})
}

// assertLoadedModel checks if the model is loaded in the container.
// For that, it checks if the response of the /api/tags endpoint
// contains the model name.
func assertLoadedModel(t *testing.T, c *ollama.OllamaContainer) {
	t.Helper()
	url, err := c.ConnectionString(context.Background())
	require.NoError(t, err)

	httpCli := &http.Client{}

	resp, err := httpCli.Get(url + "/api/tags")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	bs, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Contains(t, string(bs), "all-minilm")
}

func TestRunContainer_withModel_error(t *testing.T) {
	ctx := context.Background()

	ollamaContainer, err := ollama.Run(
		ctx,
		"ollama/ollama:0.5.7",
	)
	testcontainers.CleanupContainer(t, ollamaContainer)
	require.NoError(t, err)

	model := "non-existent"

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "pull", model})
	require.NoError(t, err)

	// we need to parse the response here to check if the error message is correct
	_, r, err := ollamaContainer.Exec(ctx, []string{"ollama", "run", model}, exec.Multiplexed())
	require.NoError(t, err)

	bs, err := io.ReadAll(r)
	require.NoError(t, err)

	stdOutput := string(bs)
	require.Contains(t, stdOutput, "Error: pull model manifest: file does not exist")
}
