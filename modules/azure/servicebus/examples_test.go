package servicebus_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/servicebus"
)

func ExampleRun() {
	ctx := context.Background()

	serviceBusContainer, err := servicebus.Run(ctx, "mcr.microsoft.com/azure-messaging/servicebus-emulator:1.0.1", servicebus.WithAcceptEULA())
	defer func() {
		if err := testcontainers.TerminateContainer(serviceBusContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := serviceBusContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

// ExampleRun_authenticateCreateClient is inspired by the example from the Azure Service Bus Go SDK:
// https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-go-how-to-use-queues?tabs=bash
func ExampleRun_authenticateCreateClient() {
	// ===== 0. Create a Service Bus config including one queue =====
	// cfg {
	cfg := `{
    "UserConfig": {
        "Namespaces": [
            {
                "Name": "sbemulatorns",
                "Queues": [
                    {
                        "Name": "queue.1",
                        "Properties": {
                            "DeadLetteringOnMessageExpiration": false,
                            "DefaultMessageTimeToLive": "PT1H",
                            "DuplicateDetectionHistoryTimeWindow": "PT20S",
                            "ForwardDeadLetteredMessagesTo": "",
                            "ForwardTo": "",
                            "LockDuration": "PT1M",
                            "MaxDeliveryCount": 10,
                            "RequiresDuplicateDetection": false,
                            "RequiresSession": false
                        }
                    }
                ]
            }
        ],
        "Logging": {
            "Type": "File"
        }
    }
}`
	// }

	// ===== 1. Run the Service Bus container =====
	// runServiceBusContainer {
	ctx := context.Background()

	serviceBusContainer, err := servicebus.Run(
		ctx,
		"mcr.microsoft.com/azure-messaging/servicebus-emulator:1.0.1",
		servicebus.WithAcceptEULA(),
		servicebus.WithConfig(strings.NewReader(cfg)),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(serviceBusContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	// ===== 2. Create a Service Bus client using a connection string to the namespace =====
	// createClient {
	client, err := azservicebus.NewClientFromConnectionString(serviceBusContainer.MustConnectionString(ctx), nil)
	if err != nil {
		log.Printf("failed to create client: %s", err)
		return
	}
	// }

	// ===== 3. Send messages to a queue =====
	// sendMessages {
	message := "Hello, Testcontainers!"

	sender, err := client.NewSender("queue.1", nil)
	if err != nil {
		log.Printf("failed to create sender: %s", err)
		return
	}
	defer sender.Close(context.TODO())

	sbMessage := &azservicebus.Message{
		Body: []byte(message),
	}
	err = sender.SendMessage(context.TODO(), sbMessage, nil)
	if err != nil {
		fmt.Printf("failed to send message: %s", err)
		return
	}
	// }

	// ===== 4. Receive messages from the queue =====
	// receiveMessages {
	receiver, err := client.NewReceiverForQueue("queue.1", nil)
	if err != nil {
		fmt.Printf("failed to create receiver: %s", err)
		return
	}
	defer receiver.Close(context.TODO())

	// Receive 1 message from the queue
	messagesCount := 1

	messages, err := receiver.ReceiveMessages(context.TODO(), messagesCount, nil)
	if err != nil {
		fmt.Printf("failed to receive messages: %s", err)
		return
	}

	fmt.Printf("received %d messages\n", len(messages))

	for _, message := range messages {
		body := message.Body
		fmt.Printf("%s\n", string(body))

		err = receiver.CompleteMessage(context.TODO(), message, nil)
		if err != nil {
			fmt.Printf("failed to complete message: %s", err)
			return
		}
	}
	// }

	// Output:
	// received 1 messages
	// Hello, Testcontainers!
}
