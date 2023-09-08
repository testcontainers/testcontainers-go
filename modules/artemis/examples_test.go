package artemis_test

import (
	"context"
	"fmt"

	"github.com/go-stomp/stomp/v3"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/artemis"
)

func ExampleRunContainer() {
	// runArtemisContainer {
	ctx := context.Background()

	artemisContainer, err := artemis.RunContainer(ctx,
		testcontainers.WithImage("docker.io/apache/activemq-artemis:2.30.0"),
		artemis.WithCredentials("test", "test"),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := artemisContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := artemisContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// connectToArtemisContainer {
	// Get broker endpoint.
	host, err := artemisContainer.BrokerEndpoint(ctx)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	defer func() {
		if err := conn.Disconnect(); err != nil {
			panic(err)
		}
	}()
	// }

	fmt.Printf("%s:%s\n", user, pass)

	// Output:
	// true
	// test:test
}
