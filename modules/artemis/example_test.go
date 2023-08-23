package artemis_test

import (
	"context"

	"github.com/go-stomp/stomp/v3"

	"github.com/testcontainers/testcontainers-go/modules/artemis"
)

func ExampleRunContainer() {
	ctx := context.Background()

	// Run container.
	container, err := artemis.RunContainer(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	// Get broker endpoint.
	host, err := container.BrokerEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	// Get credentials.
	// containerUser {
	user := container.User()
	// }
	// containerPassword {
	pass := container.Password()
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
}
