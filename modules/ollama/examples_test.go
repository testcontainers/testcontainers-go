package ollama_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/llms"
	langchainollama "github.com/tmc/langchaingo/llms/ollama"

	"github.com/testcontainers/testcontainers-go"
	tcollama "github.com/testcontainers/testcontainers-go/modules/ollama"
)

func ExampleRun() {
	// runOllamaContainer {
	ctx := context.Background()

	ollamaContainer, err := tcollama.Run(ctx, "ollama/ollama:0.5.7")
	defer func() {
		if err := testcontainers.TerminateContainer(ollamaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := ollamaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_withModel_llama2_http() {
	// withHTTPModelLlama2 {
	ctx := context.Background()

	ollamaContainer, err := tcollama.Run(ctx, "ollama/ollama:0.5.7")
	defer func() {
		if err := testcontainers.TerminateContainer(ollamaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	model := "llama2"

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "pull", model})
	if err != nil {
		log.Printf("failed to pull model %s: %s", model, err)
		return
	}

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "run", model})
	if err != nil {
		log.Printf("failed to run model %s: %s", model, err)
		return
	}

	connectionStr, err := ollamaContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	httpClient := &http.Client{}

	// generate a response
	payload := `{
	"model": "llama2",
	"prompt":"Why is the sky blue?"
}`

	req, err := http.NewRequest(http.MethodPost, connectionStr+"/api/generate", strings.NewReader(payload))
	if err != nil {
		log.Printf("failed to create request: %s", err)
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("failed to get response: %s", err)
		return
	}
	// }

	fmt.Println(resp.StatusCode)

	// Intentionally not asserting the output, as we don't want to run this example in the tests.
}

func ExampleRun_withModel_llama2_langchain() {
	// withLangchainModelLlama2 {
	ctx := context.Background()

	ollamaContainer, err := tcollama.Run(ctx, "ollama/ollama:0.5.7")
	defer func() {
		if err := testcontainers.TerminateContainer(ollamaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	model := "llama2"

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "pull", model})
	if err != nil {
		log.Printf("failed to pull model %s: %s", model, err)
		return
	}

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "run", model})
	if err != nil {
		log.Printf("failed to run model %s: %s", model, err)
		return
	}

	connectionStr, err := ollamaContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	var llm *langchainollama.LLM
	if llm, err = langchainollama.New(
		langchainollama.WithModel(model),
		langchainollama.WithServerURL(connectionStr),
	); err != nil {
		log.Printf("failed to create langchain ollama: %s", err)
		return
	}

	completion, err := llm.Call(
		context.Background(),
		"how can Testcontainers help with testing?",
		llms.WithSeed(42),         // the lower the seed, the more deterministic the completion
		llms.WithTemperature(0.0), // the lower the temperature, the more creative the completion
	)
	if err != nil {
		log.Printf("failed to create langchain ollama: %s", err)
		return
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

func ExampleRun_withLocal() {
	ctx := context.Background()

	// localOllama {
	ollamaContainer, err := tcollama.Run(ctx, "ollama/ollama:0.3.13", tcollama.WithUseLocal("OLLAMA_DEBUG=true"))
	defer func() {
		if err := testcontainers.TerminateContainer(ollamaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	model := "llama3.2:1b"

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "pull", model})
	if err != nil {
		log.Printf("failed to pull model %s: %s", model, err)
		return
	}

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "run", model})
	if err != nil {
		log.Printf("failed to run model %s: %s", model, err)
		return
	}

	connectionStr, err := ollamaContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	var llm *langchainollama.LLM
	if llm, err = langchainollama.New(
		langchainollama.WithModel(model),
		langchainollama.WithServerURL(connectionStr),
	); err != nil {
		log.Printf("failed to create langchain ollama: %s", err)
		return
	}

	completion, err := llm.Call(
		context.Background(),
		"how can Testcontainers help with testing?",
		llms.WithSeed(42),         // the lower the seed, the more deterministic the completion
		llms.WithTemperature(0.0), // the lower the temperature, the more creative the completion
	)
	if err != nil {
		log.Printf("failed to create langchain ollama: %s", err)
		return
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

	// Intentionally not asserting the output, as we don't want to run this example in the tests.
}

func ExampleRun_withImageMount() {
	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	if err != nil {
		log.Printf("failed to create docker client: %s", err)
		return
	}

	info, err := cli.Info(context.Background())
	if err != nil {
		log.Printf("failed to get docker info: %s", err)
		return
	}

	// skip if the major version of the server is not v28 or greater
	if info.ServerVersion < "28.0.0" {
		log.Printf("skipping test because the server version is not v28 or greater")
		return
	}

	ctx := context.Background()

	ollamaContainer, err := tcollama.Run(ctx, "ollama/ollama:0.5.12")
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	defer func() {
		if err := testcontainers.TerminateContainer(ollamaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	code, _, err := ollamaContainer.Exec(ctx, []string{"ollama", "pull", "all-minilm"})
	if err != nil {
		log.Printf("failed to pull model %s: %s", "all-minilm", err)
		return
	}

	fmt.Println(code)

	targetImage := "testcontainers/ollama:tc-model-all-minilm"

	err = ollamaContainer.Commit(ctx, targetImage)
	if err != nil {
		log.Printf("failed to commit container: %s", err)
		return
	}

	// start a new fresh ollama container mounting the target image
	// mountImage {
	newOllamaContainer, err := tcollama.Run(
		ctx,
		"ollama/ollama:0.5.12",
		testcontainers.WithImageMount(targetImage, "root/.ollama/models/", "/root/.ollama/models/"),
	)
	// }
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	defer func() {
		if err := testcontainers.TerminateContainer(newOllamaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	// perform an HTTP request to the ollama container to verify the model is available

	connectionStr, err := newOllamaContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	resp, err := http.Get(connectionStr + "/api/tags")
	if err != nil {
		log.Printf("failed to get request: %s", err)
		return
	}

	fmt.Println(resp.StatusCode)

	type tagsResponse struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	var tags tagsResponse
	err = json.NewDecoder(resp.Body).Decode(&tags)
	if err != nil {
		log.Printf("failed to decode response: %s", err)
		return
	}

	fmt.Println(tags.Models[0].Name)

	// Intentionally not asserting the output, as we don't want to run this example in the tests.
}
