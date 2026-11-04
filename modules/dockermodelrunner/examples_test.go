package dockermodelrunner_test

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

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
