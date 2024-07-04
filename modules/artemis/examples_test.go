package artemis_test

import (
	"context"
	"fmt"
	"log"

	"github.com/go-stomp/stomp/v3"

	"github.com/testcontainers/testcontainers-go/modules/artemis"
)

func ExampleRun() {
	// runArtemisContainer {
	ctx := context.Background()

	artemisContainer, err := artemis.Run(ctx,
		"docker.io/apache/activemq-artemis:2.30.0",
		artemis.WithCredentials("test", "test"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := artemisContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := artemisContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// connectToArtemisContainer {
	// Get broker endpoint.
	host, err := artemisContainer.BrokerEndpoint(ctx)
	if err != nil {
		log.Fatalf("failed to get broker endpoint: %s", err)
	}

	// containerUser {
	user := artemisContainer.User()
	// }
	// containerPassword {
	pass := artemisContainer.Password()
	// }

	// Connect to Artemis via STOMP.
	conn, err := stomp.Dial("tcp", host, stomp.ConnOpt.Login(user, pass))
	if err != nil {
		log.Fatalf("failed to connect to Artemis: %s", err)
	}
	defer func() {
		if err := conn.Disconnect(); err != nil {
			log.Fatalf("failed to disconnect from Artemis: %s", err)
		}
	}()
	// }

	fmt.Printf("%s:%s\n", user, pass)

	// Output:
	// true
	// test:test
}
