package dockermodelrunner_test

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/tmc/langchaingo/llms"
	langchainopenai "github.com/tmc/langchaingo/llms/openai"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner"
)

func ExampleRun_withModel() {
	ok, err := isDockerDesktopRunning(log.Default())
	if err != nil {
		log.Printf("failed to check if Docker Desktop is running: %s", err)
		return
	}
	if !ok {
		fmt.Println("true")
		return
	}

	// runWithModel {
	ctx := context.Background()

	const (
		modelNamespace = "ai"
		modelName      = "smollm2"
		modelTag       = "360M-Q4_K_M"
		fqModelName    = modelNamespace + "/" + modelName + ":" + modelTag
	)

	dmrCtr, err := dockermodelrunner.Run(
		ctx,
		dockermodelrunner.WithModel(fqModelName),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(dmrCtr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := dmrCtr.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_pullModel() {
	ok, err := isDockerDesktopRunning(log.Default())
	if err != nil {
		log.Printf("failed to check if Docker Desktop is running: %s", err)
		return
	}
	if !ok {
		fmt.Println("true")
		fmt.Println("model pulled")
		return
	}

	ctx := context.Background()

	dmrCtr, err := dockermodelrunner.Run(ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(dmrCtr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := dmrCtr.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// runPullModel {
	const (
		modelNamespace = "ai"
		modelName      = "smollm2"
		modelTag       = "360M-Q4_K_M"
		fqModelName    = modelNamespace + "/" + modelName + ":" + modelTag
	)

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err = dmrCtr.PullModel(ctx, fqModelName)
	if err != nil {
		log.Printf("failed to pull model: %s", err)
		return
	}
	// }
	fmt.Println("model pulled")

	// Output:
	// true
	// model pulled
}

func ExampleRun_inspectModel() {
	const (
		modelNamespace = "ai"
		modelName      = "smollm2"
		modelTag       = "360M-Q4_K_M"
	)

	ok, err := isDockerDesktopRunning(log.Default())
	if err != nil {
		log.Printf("failed to check if Docker Desktop is running: %s", err)
		return
	}
	if !ok {
		fmt.Println(modelNamespace + "/" + modelName + ":" + modelTag)
		return
	}

	ctx := context.Background()

	dmrCtr, err := dockermodelrunner.Run(ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(dmrCtr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// runInspectModel {
	err = dmrCtr.PullModel(ctx, modelNamespace+"/"+modelName+":"+modelTag)
	if err != nil {
		log.Printf("failed to pull model: %s", err)
		return
	}

	model, err := dmrCtr.InspectModel(ctx, modelNamespace, modelName+":"+modelTag)
	if err != nil {
		log.Printf("failed to get model: %s", err)
		return
	}
	// }
	fmt.Println(model.Tags[0])

	// Output:
	// ai/smollm2:360M-Q4_K_M
}

func ExampleRun_listModels() {
	const (
		modelNamespace = "ai"
		modelName      = "smollm2"
		modelTag       = "360M-Q4_K_M"
	)

	ok, err := isDockerDesktopRunning(log.Default())
	if err != nil {
		log.Printf("failed to check if Docker Desktop is running: %s", err)
		return
	}
	if !ok {
		fmt.Println(modelNamespace + "/" + modelName + ":" + modelTag)
		return
	}

	ctx := context.Background()

	dmrCtr, err := dockermodelrunner.Run(ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(dmrCtr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// runListModels {
	err = dmrCtr.PullModel(ctx, modelNamespace+"/"+modelName+":"+modelTag)
	if err != nil {
		log.Printf("failed to pull model: %s", err)
		return
	}

	models, err := dmrCtr.ListModels(ctx)
	if err != nil {
		log.Printf("failed to get model: %s", err)
		return
	}
	// }
	for _, model := range models {
		if slices.Contains(model.Tags, modelNamespace+"/"+modelName+":"+modelTag) {
			fmt.Println(model.Tags[0])
		}
	}

	// Output:
	// ai/smollm2:360M-Q4_K_M
}

func ExampleRun_openAI() {
	ok, err := isDockerDesktopRunning(log.Default())
	if err != nil {
		log.Printf("failed to check if Docker Desktop is running: %s", err)
		return
	}
	if !ok {
		fmt.Println("true")
		return
	}

	ctx := context.Background()

	dmrCtr, err := dockermodelrunner.Run(ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(dmrCtr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	const (
		modelNamespace = "ai"
		modelName      = "smollm2"
		modelTag       = "360M-Q4_K_M"
	)

	err = dmrCtr.PullModel(ctx, modelNamespace+"/"+modelName+":"+modelTag)
	if err != nil {
		log.Printf("failed to pull model: %s", err)
		return
	}

	llmURL := dmrCtr.OpenAIEndpoint()

	client := openai.NewClient(
		option.WithBaseURL(llmURL),
		option.WithAPIKey(""), // No API key needed for Model Runner
	)

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a useful AI agent expert with TV series."),
		openai.UserMessage("Tell me about the Anime series called Attack on Titan?"),
	}

	param := openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       modelNamespace + "/" + modelName + ":" + modelTag,
		Temperature: openai.Opt(0.8),
		MaxTokens:   openai.Opt(int64(1024)),
	}

	completion, err := client.Chat.Completions.New(ctx, param)
	if err != nil {
		log.Println("😡:", err)
		return
	}

	log.Println(completion.Choices[0].Message.Content)
	fmt.Println(len(completion.Choices[0].Message.Content) > 0)

	// Output:
	// true
}

func ExampleRun_langchaingo() {
	ok, err := isDockerDesktopRunning(log.Default())
	if err != nil {
		log.Printf("failed to check if Docker Desktop is running: %s", err)
		return
	}
	if !ok {
		fmt.Println("true")
		return
	}

	ctx := context.Background()

	dmrCtr, err := dockermodelrunner.Run(ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(dmrCtr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	const (
		modelNamespace = "ai"
		modelName      = "smollm2"
		modelTag       = "360M-Q4_K_M"
	)

	err = dmrCtr.PullModel(ctx, modelNamespace+"/"+modelName+":"+modelTag)
	if err != nil {
		log.Printf("failed to pull model: %s", err)
		return
	}

	llmURL := dmrCtr.OpenAIEndpoint()

	opts := []langchainopenai.Option{
		langchainopenai.WithBaseURL(llmURL),
		langchainopenai.WithModel(modelNamespace + "/" + modelName + ":" + modelTag),
		langchainopenai.WithToken("foo"), // No API key needed for Model Runner
	}

	llm, err := langchainopenai.New(opts...)
	if err != nil {
		log.Println(err)
		return
	}

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a useful AI agent expert with TV series."),
		llms.TextParts(llms.ChatMessageTypeHuman, "Tell me about the Anime series called Attack on Titan"),
	}

	var streamingStrings []string
	_, err = llm.GenerateContent(ctx, content,
		llms.WithMaxTokens(1024),
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			streamingStrings = append(streamingStrings, string(chunk))
			return nil
		}))
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(strings.Join(streamingStrings, ""))
	fmt.Println(len(streamingStrings) > 0)

	// Output:
	// true
}
