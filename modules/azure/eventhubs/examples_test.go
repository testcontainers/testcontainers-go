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
	"github.com/testcontainers/testcontainers-go/modules/azure/eventhubs"
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
