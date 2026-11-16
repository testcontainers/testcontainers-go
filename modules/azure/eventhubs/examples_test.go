package eventhubs_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/azurite"
	"github.com/testcontainers/testcontainers-go/modules/azure/eventhubs"
	"github.com/testcontainers/testcontainers-go/network"
)

func ExampleRun() {
	ctx := context.Background()

	eventHubsCtr, err := eventhubs.Run(ctx, "mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0", eventhubs.WithAcceptEULA())
	defer func() {
		if err := testcontainers.TerminateContainer(eventHubsCtr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := eventHubsCtr.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

// ExampleRun_withAzuriteContainer demonstrates how to wire in a pre-existing
// Azurite container so that the Event Hubs emulator shares it. The caller is
// responsible for tearing down Azurite and the network; the Event Hubs
// container will not touch them when Terminate is called.
func ExampleRun_withAzuriteContainer() {
	ctx := context.Background()

	// withAzuriteContainer_network {
	nw, err := network.New(ctx)
	if err != nil {
		log.Printf("failed to create network: %s", err)
		return
	}
	defer func() {
		if err := nw.Remove(ctx); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()
	// }

	// withAzuriteContainer_azurite {
	azuriteCtr, err := azurite.Run(
		ctx,
		"mcr.microsoft.com/azure-storage/azurite:3.36.0",
		network.WithNetwork([]string{"azurite"}, nw),
		testcontainers.WithEntrypointArgs("--skipApiVersionCheck"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(azuriteCtr); err != nil {
			log.Printf("failed to terminate azurite container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start azurite container: %s", err)
		return
	}
	// }

	// withAzuriteContainer_eventhubs {
	eventHubsCtr, err := eventhubs.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
		eventhubs.WithAcceptEULA(),
		eventhubs.WithAzuriteContainer(azuriteCtr, nw, "azurite"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(eventHubsCtr); err != nil {
			log.Printf("failed to terminate eventhubs container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start eventhubs container: %s", err)
		return
	}
	// }

	state, err := eventHubsCtr.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

// ExampleRun_withConfigObject demonstrates how to build a statically-typed
// Event Hubs emulator configuration using the functional-options API and
// inject it into the container via WithConfigObject.
func ExampleRun_withConfigObject() {
	ctx := context.Background()

	// withConfigObject_buildConfig {
	cfg, err := eventhubs.NewConfig(
		eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
			eventhubs.WithEntity("eh1", 2,
				eventhubs.WithConsumerGroup("cg1"),
				eventhubs.WithConsumerGroup("$Default"),
			),
			eventhubs.WithEntity("eh2", 1,
				eventhubs.WithConsumerGroup("cg1"),
			),
		),
	)
	if err != nil {
		log.Printf("failed to build eventhubs config: %s", err)
		return
	}
	// }

	// withConfigObject_run {
	eventHubsCtr, err := eventhubs.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
		eventhubs.WithAcceptEULA(),
		eventhubs.WithConfigObject(cfg),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(eventHubsCtr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := eventHubsCtr.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)
}

// ExampleNewConfig demonstrates how to construct an Event Hubs emulator
// configuration using the three-level functional-options API without starting
// any containers.
func ExampleNewConfig() {
	// ExampleNewConfig_build {
	cfg, err := eventhubs.NewConfig(
		eventhubs.WithLoggingType("File"),
		eventhubs.WithNamespace(eventhubs.EmulatorNamespaceName,
			eventhubs.WithEntity("eh1", 1,
				eventhubs.WithConsumerGroup("cg1"),
			),
		),
	)
	if err != nil {
		log.Printf("failed to build config: %s", err)
		return
	}
	// }

	fmt.Println(cfg.UserConfig.LoggingConfig.Type)
	fmt.Println(cfg.UserConfig.NamespaceConfig[0].Name)

	// Output:
	// File
	// emulatorns1
}
