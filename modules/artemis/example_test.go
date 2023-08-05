package artemis_test

import (
	"context"

	"github.com/go-stomp/stomp/v3"
	"github.com/testcontainers/testcontainers-go/modules/artemis"
)

func ExampleRunContainer() {
	ctx := context.Background()

	user := "username"
	pass := "password"

	// runContainer {
	container, err := artemis.RunContainer(ctx, artemis.WithCredentials(user, pass))
	if err != nil {
		panic(err)
	}
	// }

	host, err := container.BrokerEndpoint(ctx)
	if err != nil {
		panic(err)
	}

	conn, err := stomp.Dial("tcp", host, stomp.ConnOpt.Login(user, pass))
	if err != nil {
		panic(err)
	}

	_ = conn
}
