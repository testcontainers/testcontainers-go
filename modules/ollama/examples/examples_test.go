package examples_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms"
	langchainollama "github.com/tmc/langchaingo/llms/ollama"

	"github.com/testcontainers/testcontainers-go"
	tcollama "github.com/testcontainers/testcontainers-go/modules/ollama"
)

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
