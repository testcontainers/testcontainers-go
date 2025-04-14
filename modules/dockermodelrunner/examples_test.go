package dockermodelrunner_test

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner"
)

func ExampleRun_withModel() {
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
		"alpine/socat:1.8.0.1",
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
	ctx := context.Background()

	dmrCtr, err := dockermodelrunner.Run(
		ctx,
		"alpine/socat:1.8.0.1",
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

func ExampleRun_getModel() {
	ctx := context.Background()

	dmrCtr, err := dockermodelrunner.Run(
		ctx,
		"alpine/socat:1.8.0.1",
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

	// runGetModel {
	const (
		modelNamespace = "ai"
		modelName      = "smollm2"
	)

	err = dmrCtr.PullModel(ctx, modelNamespace+"/"+modelName)
	if err != nil {
		log.Printf("failed to pull model: %s", err)
		return
	}

	model, err := dmrCtr.GetModel(ctx, modelNamespace, modelName)
	if err != nil {
		log.Printf("failed to get model: %s", err)
		return
	}
	// }
	fmt.Println(model.Tags[0])
	fmt.Println(model.Tags[1])

	// Output:
	// ai/smollm2
	// ai/smollm2:360M-Q4_K_M
}

func ExampleRun_listModels() {
	ctx := context.Background()

	dmrCtr, err := dockermodelrunner.Run(
		ctx,
		"alpine/socat:1.8.0.1",
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

	// runListModels {
	const (
		modelNamespace = "ai"
		modelName      = "smollm2"
	)

	err = dmrCtr.PullModel(ctx, modelNamespace+"/"+modelName)
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
		if slices.Contains(model.Tags, modelNamespace+"/"+modelName) {
			fmt.Println(model.Tags[0])
			fmt.Println(model.Tags[1])
		}
	}

	// Output:
	// ai/smollm2
	// ai/smollm2:360M-Q4_K_M
}

func ExampleRun_openAI() {
	ctx := context.Background()

	dmrCtr, err := dockermodelrunner.Run(
		ctx,
		"alpine/socat:1.8.0.1",
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

	const (
		modelNamespace = "ai"
		modelName      = "smollm2"
	)

	err = dmrCtr.PullModel(ctx, modelNamespace+"/"+modelName)
	if err != nil {
		log.Printf("failed to pull model: %s", err)
		return
	}

	llmURL := dmrCtr.OpenAIEndpoint(ctx)

	client := openai.NewClient(
		option.WithBaseURL(llmURL),
		option.WithAPIKey(""), // No API key needed for Model Runner
	)

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a useful AI agent expert with TV series."),
		openai.UserMessage("Tell me about the English series called The Avengers?"),
	}

	param := openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       modelNamespace + "/" + modelName,
		Temperature: openai.Opt(0.8),
	}

	completion, err := client.Chat.Completions.New(ctx, param)
	if err != nil {
		log.Fatalln("😡:", err)
	}

	log.Println(completion.Choices[0].Message.Content)
	fmt.Println(len(completion.Choices[0].Message.Content) > 0)

	// Output:
	// true
}
