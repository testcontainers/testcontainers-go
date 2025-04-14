package dockermodelrunner_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner"
)

func ExampleRun_pullModel() {
	ctx := context.Background()

	const (
		modelNamespace = "ai"
		modelName      = "llama3.2"
		modelTag       = "latest"
		fqModelName    = modelNamespace + "/" + modelName + ":" + modelTag
	)

	dockermodelrunnerContainer, err := dockermodelrunner.Run(
		ctx,
		"alpine/socat:1.8.0.1",
		dockermodelrunner.WithModel(fqModelName),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(dockermodelrunnerContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := dockermodelrunnerContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
