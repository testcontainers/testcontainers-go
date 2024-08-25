package ollama_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/llms"
	langchainollama "github.com/tmc/langchaingo/llms/ollama"

	tcollama "github.com/testcontainers/testcontainers-go/modules/ollama"
)

func ExampleRun() {
	// runOllamaContainer {
	ctx := context.Background()

	ollamaContainer, err := tcollama.Run(ctx, "ollama/ollama:0.1.25")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := ollamaContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := ollamaContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_withModel_llama2_http() {
	// withHTTPModelLlama2 {
	ctx := context.Background()

	ollamaContainer, err := tcollama.Run(ctx, "ollama/ollama:0.1.25")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := ollamaContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	model := "llama2"

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "pull", model})
	if err != nil {
		log.Fatalf("failed to pull model %s: %s", model, err) // nolint:gocritic
	}

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "run", model})
	if err != nil {
		log.Fatalf("failed to run model %s: %s", model, err) // nolint:gocritic
	}

	connectionStr, err := ollamaContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err) // nolint:gocritic
	}

	httpClient := &http.Client{}

	// generate a response
	payload := `{
	"model": "llama2",
	"prompt":"Why is the sky blue?"
}`

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/generate", connectionStr), strings.NewReader(payload))
	if err != nil {
		log.Fatalf("failed to create request: %s", err) // nolint:gocritic
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("failed to get response: %s", err) // nolint:gocritic
	}
	// }

	fmt.Println(resp.StatusCode)

	// Intentionally not asserting the output, as we don't want to run this example in the tests.
}

func ExampleRun_withModel_llama2_langchain() {
	// withLangchainModelLlama2 {
	ctx := context.Background()

	ollamaContainer, err := tcollama.Run(ctx, "ollama/ollama:0.1.25")
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := ollamaContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	model := "llama2"

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "pull", model})
	if err != nil {
		log.Fatalf("failed to pull model %s: %s", model, err) // nolint:gocritic
	}

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "run", model})
	if err != nil {
		log.Fatalf("failed to run model %s: %s", model, err) // nolint:gocritic
	}

	connectionStr, err := ollamaContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err) // nolint:gocritic
	}

	var llm *langchainollama.LLM
	if llm, err = langchainollama.New(
		langchainollama.WithModel(model),
		langchainollama.WithServerURL(connectionStr),
	); err != nil {
		log.Fatalf("failed to create langchain ollama: %s", err) // nolint:gocritic
	}

	completion, err := llm.Call(
		context.Background(),
		"how can Testcontainers help with testing?",
		llms.WithSeed(42),         // the lower the seed, the more deterministic the completion
		llms.WithTemperature(0.0), // the lower the temperature, the more creative the completion
	)
	if err != nil {
		log.Fatalf("failed to create langchain ollama: %s", err) // nolint:gocritic
	}

	words := []string{
		"easy", "isolation", "consistency",
	}
	lwCompletion := strings.ToLower(completion)

	for _, word := range words {
		if strings.Contains(lwCompletion, word) {
			fmt.Println(true)
		}
	}

	// }

	// Intentionally not asserting the output, as we don't want to run this example in the tests.
}
