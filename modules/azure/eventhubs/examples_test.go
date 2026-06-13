// Examples in this file has been taken from the following article:
// https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-go-get-started-send

package eventhubs_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"

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

func ExampleRun_sendEventsToEventHub() {
	ctx := context.Background()

	// cfg {
	cfg := `{
    "UserConfig": {
        "NamespaceConfig": [
            {
                "Type": "EventHub",
                "Name": "emulatorNs1",
                "Entities": [
                    {
                        "Name": "eh1",
                        "PartitionCount": "1",
                        "ConsumerGroups": [
                            {
                                "Name": "cg1"
                            }
                        ]
                    }
                ]
            }
        ],
        "LoggingConfig": {
            "Type": "File"
        }
    }
}
`
	// }

	// runEventHubsContainer {
	eventHubsCtr, err := eventhubs.Run(ctx, "mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0", eventhubs.WithAcceptEULA(), eventhubs.WithConfig(strings.NewReader(cfg)))
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

	// ===== 1. Create an Event Hubs producer client using a connection string to the namespace and the event hub =====
	// createProducerClient {
	connectionString, err := eventHubsCtr.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	producerClient, err := azeventhubs.NewProducerClientFromConnectionString(connectionString, "eh1", nil)
	if err != nil {
		log.Printf("failed to create producer client: %s", err)
		return
	}
	defer producerClient.Close(context.TODO())
	// }

	// ===== 2. Create sample events =====
	// createSampleEvents {
	events := []*azeventhubs.EventData{
		{
			Body: []byte("hello"),
		},
		{
			Body: []byte("world"),
		},
	}
	// }

	// ===== 3. Create a batch object and add sample events to the batch =====
	// createBatch {
	newBatchOptions := &azeventhubs.EventDataBatchOptions{}

	var batch *azeventhubs.EventDataBatch
	maxRetries := 3
	// Retry creating the event data batch 3 times, because the event hub is created from the configuration
	// and Testcontainers cannot add a wait strategy for the event hub to be created.
	for retries := 0; retries < maxRetries; retries++ {
		batch, err = producerClient.NewEventDataBatch(context.TODO(), newBatchOptions)
		if err == nil {
			break
		}

		if retries == maxRetries-1 {
			log.Printf("failed to create event data batch after %d attempts: %s", maxRetries, err)
			return
		}
	}

	for i := range events {
		err = batch.AddEventData(events[i], nil)
		if err != nil {
			log.Printf("failed to add event data to batch: %s", err)
			return
		}
	}
	// }

	// ===== 4. Send the batch of events to the event hub =====
	// sendEventDataBatch {
	err = producerClient.SendEventDataBatch(context.TODO(), batch, nil)
	if err != nil {
		log.Printf("failed to send event data batch: %s", err)
		return
	}
	// }

	fmt.Println(err)

	// Output:
	// <nil>
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
		"mcr.microsoft.com/azure-storage/azurite:3.33.0",
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
