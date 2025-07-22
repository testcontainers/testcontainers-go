package solace_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"

	"github.com/testcontainers/testcontainers-go"
	sc "github.com/testcontainers/testcontainers-go/modules/solace"
)

func ExampleRun() {
	ctx := context.Background()
	ctr, err := sc.Run(ctx, "solace-pubsub-standard:latest",
		sc.WithCredentials("admin", "admin"),
		sc.WithExposedPorts("5672/tcp", "8080/tcp"),
		sc.WithEnv(map[string]string{
			"username_admin_globalaccesslevel": "admin",
			"username_admin_password":          "admin",
		}),
		sc.WithShmSize(1<<30),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	fmt.Println(err)

	// Output:
	// <nil>
}

func ExampleRun_withTopicAndQueue() {
	ctx := context.Background()

	ctr, err := sc.Run(ctx, "solace-pubsub-standard:latest",
		sc.WithCredentials("admin", "admin"),
		sc.WithVpn("test-vpn"),
		sc.WithExposedPorts("5672/tcp", "8080/tcp"),
		sc.WithEnv(map[string]string{
			"username_admin_globalaccesslevel": "admin",
			"username_admin_password":          "admin",
		}),
		sc.WithShmSize(1<<30),
		sc.WithQueue("TestQueue", "Topic/MyTopic"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	fmt.Println(err)
	err = testMessagePublishAndConsume(ctr, "TestQueue", "Topic/MyTopic")
	fmt.Println(err)

	// Output:
	// <nil>
	// Published message to topic: Topic/MyTopic
	// Received message: Hello from Solace testcontainers!
	// Successfully received message from queue: TestQueue
	// <nil>
}

func testMessagePublishAndConsume(ctr *sc.SolaceContainer, queueName, topicName string) error {
	// Get the SMF service URL from the container
	smfURL, err := ctr.BrokerURLFor(context.Background(), sc.ServiceSMF)
	if err != nil {
		return fmt.Errorf("failed to get SMF URL: %w", err)
	}

	// Configure connection properties
	brokerConfig := config.ServicePropertyMap{
		config.TransportLayerPropertyHost:                 smfURL,
		config.ServicePropertyVPNName:                     ctr.Vpn(),
		config.AuthenticationPropertyScheme:               config.AuthenticationSchemeBasic,
		config.AuthenticationPropertySchemeBasicUserName:  ctr.Username(),
		config.AuthenticationPropertySchemeBasicPassword:  ctr.Password(),
		config.TransportLayerPropertyReconnectionAttempts: 0,
	}

	// Build messaging service
	messagingService, err := messaging.NewMessagingServiceBuilder().
		FromConfigurationProvider(brokerConfig).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build messaging service: %w", err)
	}

	// Connect to the messaging service
	if err := messagingService.Connect(); err != nil {
		return fmt.Errorf("failed to connect to messaging service: %w", err)
	}
	defer func() {
		if err := messagingService.Disconnect(); err != nil {
			log.Printf("Error disconnecting from messaging service: %v", err)
		}
	}()

	// Test message publishing
	err = publishTestMessage(messagingService, topicName)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait a moment for message to be delivered to queue
	time.Sleep(1 * time.Second)

	// Test message consumption from queue
	err = consumeTestMessage(messagingService, queueName)
	if err != nil {
		return fmt.Errorf("failed to consume message: %w", err)
	}

	return nil
}

func publishTestMessage(messagingService solace.MessagingService, topicName string) error {
	// Build a direct message publisher
	directPublisher, err := messagingService.CreateDirectMessagePublisherBuilder().Build()
	if err != nil {
		return fmt.Errorf("failed to build publisher: %w", err)
	}

	// Start the publisher
	if err := directPublisher.Start(); err != nil {
		return fmt.Errorf("failed to start publisher: %w", err)
	}
	defer func() {
		if err := directPublisher.Terminate(1 * time.Second); err != nil {
			log.Printf("Error terminating direct publisher: %v", err)
		}
	}()

	// Create a message
	messageBuilder := messagingService.MessageBuilder()
	message, err := messageBuilder.
		WithProperty("custom-property", "test-value").
		BuildWithStringPayload("Hello from Solace testcontainers!")
	if err != nil {
		return fmt.Errorf("failed to build message: %w", err)
	}

	// Create topic resource
	topic := resource.TopicOf(topicName)

	// Publish the message
	if err := directPublisher.Publish(message, topic); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	fmt.Printf("Published message to topic: %s\n", topicName)
	return nil
}

func consumeTestMessage(messagingService solace.MessagingService, queueName string) error {
	// Build a persistent message receiver
	persistentReceiver, err := messagingService.CreatePersistentMessageReceiverBuilder().
		Build(resource.QueueDurableExclusive(queueName))
	if err != nil {
		return fmt.Errorf("failed to build receiver: %w", err)
	}

	// Set up message handler
	messageReceived := make(chan message.InboundMessage, 1)
	errorChan := make(chan error, 1)

	messageHandler := func(msg message.InboundMessage) {
		payload, ok := msg.GetPayloadAsString()
		if ok {
			fmt.Printf("Received message: %s\n", payload)
		}
		messageReceived <- msg
	}

	// Start the receiver
	if err := persistentReceiver.Start(); err != nil {
		return fmt.Errorf("failed to start receiver: %w", err)
	}
	defer func() {
		if err := persistentReceiver.Terminate(1 * time.Second); err != nil {
			log.Printf("Error terminating persistent receiver: %v", err)
		}
	}()

	// Receive messages asynchronously
	if err := persistentReceiver.ReceiveAsync(messageHandler); err != nil {
		return fmt.Errorf("failed to start async receive: %w", err)
	}

	// Wait for message with timeout
	select {
	case <-messageReceived:
		fmt.Printf("Successfully received message from queue: %s\n", queueName)
		// For persistent messages, acknowledgment is typically handled automatically
		// or through the receiver's configuration
		return nil
	case err := <-errorChan:
		return fmt.Errorf("error receiving message: %w", err)
	case <-time.After(15 * time.Second):
		return fmt.Errorf("timeout waiting for message from queue: %s", queueName)
	}
}
