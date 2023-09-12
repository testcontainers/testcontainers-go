package nats_test

import (
	"context"
	"fmt"

	natsgo "github.com/nats-io/nats.go"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/nats"
)

func ExampleRunContainer() {
	// runNATSContainer {
	ctx := context.Background()

	natsContainer, err := nats.RunContainer(ctx,
		testcontainers.WithImage("nats:2.9"),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := natsContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := natsContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_connectWithCredentials() {
	// natsConnect {
	ctx := context.Background()

	container, err := nats.RunContainer(ctx, nats.WithUsername("foo"), nats.WithPassword("bar"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	nc, err := natsgo.Connect(uri, natsgo.UserInfo(container.User, container.Password))
	if err != nil {
		panic(err)
	}
	defer nc.Close()
	// }

	fmt.Println(nc.IsConnected())

	// Output:
	// true
}
