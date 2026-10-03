package activemq_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/activemq"
)

// ExampleRun demonstrates how to start an Apache ActiveMQ Classic container
// and retrieve the broker URL and web-console credentials.
func ExampleRun() {
	// runActiveMQContainer {
	ctx := context.Background()

	activemqContainer, err := activemq.Run(ctx,
		"apache/activemq-classic:5.18.7",
		activemq.WithAdminCredentials("test", "test"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(activemqContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := activemqContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// adminUser {
	user := activemqContainer.AdminUser()
	// }
	// adminPassword {
	pass := activemqContainer.AdminPassword()
	// }

	fmt.Printf("%s:%s\n", user, pass)

	// Output:
	// true
	// admin:admin
}
