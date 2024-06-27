package ollama_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/ollama"
)

func TestOllama(t *testing.T) {
	ctx := context.Background()

	container, err := ollama.Run(ctx, "ollama/ollama:0.1.25")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	t.Run("ConnectionString", func(t *testing.T) {
		// connectionString {
		connectionStr, err := container.ConnectionString(ctx)
		// }
		if err != nil {
			t.Fatal(err)
		}

		httpClient := &http.Client{}
		resp, err := httpClient.Get(connectionStr)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Pull and Run Model", func(t *testing.T) {
		model := "all-minilm"

		_, _, err = container.Exec(context.Background(), []string{"ollama", "pull", model})
		if err != nil {
			log.Fatalf("failed to pull model %s: %s", model, err)
		}

		_, _, err = container.Exec(context.Background(), []string{"ollama", "run", model})
		if err != nil {
			log.Fatalf("failed to run model %s: %s", model, err)
		}

		assertLoadedModel(t, container)
	})

	t.Run("Commit to image including model", func(t *testing.T) {
		// commitOllamaContainer {

		// Defining the target image name based on the default image and a random string.
		// Users can change the way this is generated, but it should be unique.
		targetImage := fmt.Sprintf("%s-%s", ollama.DefaultOllamaImage, strings.ToLower(uuid.New().String()[:4]))

		err := container.Commit(context.Background(), targetImage)
		// }
		if err != nil {
			t.Fatal(err)
		}

		newOllamaContainer, err := ollama.Run(
			context.Background(),
			targetImage,
		)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := newOllamaContainer.Terminate(context.Background()); err != nil {
				t.Fatalf("failed to terminate container: %s", err)
			}
		})

		assertLoadedModel(t, newOllamaContainer)
	})
}

// assertLoadedModel checks if the model is loaded in the container.
// For that, it checks if the response of the /api/tags endpoint
// contains the model name.
func assertLoadedModel(t *testing.T, c *ollama.OllamaContainer) {
	url, err := c.ConnectionString(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	httpCli := &http.Client{}

	resp, err := httpCli.Get(url + "/api/tags")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", resp.StatusCode)
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(bs), "all-minilm") {
		t.Fatalf("expected response to contain all-minilm, got %s", string(bs))
	}
}

func TestRunContainer_withModel_error(t *testing.T) {
	ctx := context.Background()

	ollamaContainer, err := ollama.Run(
		ctx,
		"ollama/ollama:0.1.25",
	)
	if err != nil {
		t.Fatalf("expected error to be nil, got %s", err)
	}

	model := "non-existent"

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "pull", model})
	if err != nil {
		log.Fatalf("expected nil error, got %s", err)
	}

	// we need to parse the response here to check if the error message is correct
	_, r, err := ollamaContainer.Exec(ctx, []string{"ollama", "run", model}, exec.Multiplexed())
	if err != nil {
		log.Fatalf("expected nil error, got %s", err)
	}

	bs, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to run %s model: %s", model, err)
	}

	stdOutput := string(bs)
	if !strings.Contains(stdOutput, "Error: pull model manifest: file does not exist") {
		t.Fatalf("expected output to contain %q, got %s", "Error: pull model manifest: file does not exist", stdOutput)
	}
}
