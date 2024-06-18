package nats_test

import (
	"context"
	"fmt"
	"log"
	"time"

	natsgo "github.com/nats-io/nats.go"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/nats"
	"github.com/testcontainers/testcontainers-go/network"
)

func ExampleRunContainer() {
	// runNATSContainer {
	ctx := context.Background()

	natsContainer, err := nats.RunContainer(ctx,
		testcontainers.WithImage("nats:2.9"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := natsContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := natsContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
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
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err) // nolint:gocritic
	}

	nc, err := natsgo.Connect(uri, natsgo.UserInfo(container.User, container.Password))
	if err != nil {
		log.Fatalf("failed to connect to NATS: %s", err)
	}
	defer nc.Close()
	// }

	fmt.Println(nc.IsConnected())

	// Output:
	// true
}

func ExampleRunContainer_cluster() {
	ctx := context.Background()

	nwr, err := network.New(ctx)
	if err != nil {
		log.Fatalf("failed to create network: %s", err)
	}

	// withArguments {
	natsContainer1, err := nats.RunContainer(ctx,
		network.WithNetwork([]string{"nats1"}, nwr),
		nats.WithArgument("name", "nats1"),
		nats.WithArgument("cluster_name", "c1"),
		nats.WithArgument("cluster", "nats://nats1:6222"),
		nats.WithArgument("routes", "nats://nats1:6222,nats://nats2:6222,nats://nats3:6222"),
		nats.WithArgument("http_port", "8222"),
	)
	// }
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	// Clean up the container
	defer func() {
		if err := natsContainer1.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	natsContainer2, err := nats.RunContainer(ctx,
		network.WithNetwork([]string{"nats2"}, nwr),
		nats.WithArgument("name", "nats2"),
		nats.WithArgument("cluster_name", "c1"),
		nats.WithArgument("cluster", "nats://nats2:6222"),
		nats.WithArgument("routes", "nats://nats1:6222,nats://nats2:6222,nats://nats3:6222"),
		nats.WithArgument("http_port", "8222"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err) // nolint:gocritic
	}
	// Clean up the container
	defer func() {
		if err := natsContainer2.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	natsContainer3, err := nats.RunContainer(ctx,
		network.WithNetwork([]string{"nats3"}, nwr),
		nats.WithArgument("name", "nats3"),
		nats.WithArgument("cluster_name", "c1"),
		nats.WithArgument("cluster", "nats://nats3:6222"),
		nats.WithArgument("routes", "nats://nats1:6222,nats://nats2:6222,nats://nats3:6222"),
		nats.WithArgument("http_port", "8222"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err) // nolint:gocritic
	}
	defer func() {
		if err := natsContainer3.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// cluster URL
	servers := natsContainer1.MustConnectionString(ctx) + "," + natsContainer2.MustConnectionString(ctx) + "," + natsContainer3.MustConnectionString(ctx)

	nc, err := natsgo.Connect(servers, natsgo.MaxReconnects(5), natsgo.ReconnectWait(2*time.Second))
	if err != nil {
		log.Fatalf("connecting to nats container failed:\n\t%v\n", err) // nolint:gocritic
	}

	{
		// Simple Publisher
		err = nc.Publish("foo", []byte("Hello World"))
		if err != nil {
			log.Fatalf("failed to publish message: %s", err) // nolint:gocritic
		}
	}

	{
		// Channel subscriber
		ch := make(chan *natsgo.Msg, 64)
		sub, err := nc.ChanSubscribe("channel", ch)
		if err != nil {
			log.Fatalf("failed to subscribe to message: %s", err) // nolint:gocritic
		}

		// Request
		err = nc.Publish("channel", []byte("Hello NATS Cluster!"))
		if err != nil {
			log.Fatalf("failed to publish message: %s", err) // nolint:gocritic
		}

		msg := <-ch
		fmt.Println(string(msg.Data))

		err = sub.Unsubscribe()
		if err != nil {
			log.Fatalf("failed to unsubscribe: %s", err) // nolint:gocritic
		}

		err = sub.Drain()
		if err != nil {
			log.Fatalf("failed to drain: %s", err) // nolint:gocritic
		}
	}

	{
		// Responding to a request message
		sub, err := nc.Subscribe("request", func(m *natsgo.Msg) {
			err1 := m.Respond([]byte("answer is 42"))
			if err1 != nil {
				log.Fatalf("failed to respond to message: %s", err1) // nolint:gocritic
			}
		})
		if err != nil {
			log.Fatalf("failed to subscribe to message: %s", err) // nolint:gocritic
		}

		// Request
		msg, err := nc.Request("request", []byte("what is the answer?"), 1*time.Second)
		if err != nil {
			log.Fatalf("failed to send request: %s", err) // nolint:gocritic
		}

		fmt.Println(string(msg.Data))

		err = sub.Unsubscribe()
		if err != nil {
			log.Fatalf("failed to unsubscribe: %s", err) // nolint:gocritic
		}

		err = sub.Drain()
		if err != nil {
			log.Fatalf("failed to drain: %s", err) // nolint:gocritic
		}
	}

	// Drain connection (Preferred for responders)
	// Close() not needed if this is called.
	err = nc.Drain()
	if err != nil {
		log.Fatalf("failed to drain connection: %s", err) // nolint:gocritic
	}

	// Close connection
	nc.Close()

	// Output:
	// Hello NATS Cluster!
	// answer is 42
}
