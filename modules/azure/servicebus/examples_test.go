package servicebus_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/servicebus"
)

func ExampleRun() {
	ctx := context.Background()

	serviceBusContainer, err := servicebus.Run(ctx, "mcr.microsoft.com/azure-messaging/servicebus-emulator:1.1.2", servicebus.WithAcceptEULA())
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
