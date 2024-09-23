package artemis_test

import (
	"context"
	"fmt"
	"log"

	"github.com/go-stomp/stomp/v3"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/artemis"
)

func ExampleRun() {
	// runArtemisContainer {
	ctx := context.Background()

	artemisContainer, err := artemis.Run(ctx,
		"docker.io/apache/activemq-artemis:2.30.0",
		artemis.WithCredentials("test", "test"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(artemisContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := artemisContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// connectToArtemisContainer {
	// Get broker endpoint.
	host, err := artemisContainer.BrokerEndpoint(ctx)
	if err != nil {
		log.Printf("failed to get broker endpoint: %s", err)
		return
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
		log.Printf("failed to connect to Artemis: %s", err)
		return
	}
	defer func() {
		if err := conn.Disconnect(); err != nil {
			log.Printf("failed to disconnect from Artemis: %s", err)
		}
	}()
	// }

	fmt.Printf("%s:%s\n", user, pass)

	// Output:
	// true
	// test:test
}
