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

func ExampleRun() {
	// runNATSContainer {
	ctx := context.Background()

	natsContainer, err := nats.Run(ctx, "nats:2.9")
	defer func() {
		if err := testcontainers.TerminateContainer(natsContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := natsContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connectWithCredentials() {
	// natsConnect {
	ctx := context.Background()

	ctr, err := nats.Run(ctx, "nats:2.9", nats.WithUsername("foo"), nats.WithPassword("bar"))
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	uri, err := ctr.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	nc, err := natsgo.Connect(uri, natsgo.UserInfo(ctr.User, ctr.Password))
	if err != nil {
		log.Printf("failed to connect to NATS: %s", err)
		return
	}
	defer nc.Close()
	// }

	fmt.Println(nc.IsConnected())

	// Output:
	// true
}

func ExampleRun_cluster() {
	ctx := context.Background()

	nwr, err := network.New(ctx)
	if err != nil {
		log.Printf("failed to create network: %s", err)
		return
	}

	defer func() {
		if err := nwr.Remove(context.Background()); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()

	// withArguments {
	natsContainer1, err := nats.Run(ctx,
		"nats:2.9",
		network.WithNetwork([]string{"nats1"}, nwr),
		nats.WithArgument("name", "nats1"),
		nats.WithArgument("cluster_name", "c1"),
		nats.WithArgument("cluster", "nats://nats1:6222"),
		nats.WithArgument("routes", "nats://nats1:6222,nats://nats2:6222,nats://nats3:6222"),
		nats.WithArgument("http_port", "8222"),
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(natsContainer1); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	natsContainer2, err := nats.Run(ctx,
		"nats:2.9",
		network.WithNetwork([]string{"nats2"}, nwr),
		nats.WithArgument("name", "nats2"),
		nats.WithArgument("cluster_name", "c1"),
		nats.WithArgument("cluster", "nats://nats2:6222"),
		nats.WithArgument("routes", "nats://nats1:6222,nats://nats2:6222,nats://nats3:6222"),
		nats.WithArgument("http_port", "8222"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(natsContainer2); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	natsContainer3, err := nats.Run(ctx,
		"nats:2.9",
		network.WithNetwork([]string{"nats3"}, nwr),
		nats.WithArgument("name", "nats3"),
		nats.WithArgument("cluster_name", "c1"),
		nats.WithArgument("cluster", "nats://nats3:6222"),
		nats.WithArgument("routes", "nats://nats1:6222,nats://nats2:6222,nats://nats3:6222"),
		nats.WithArgument("http_port", "8222"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(natsContainer3); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// cluster URL
	servers := natsContainer1.MustConnectionString(ctx) + "," + natsContainer2.MustConnectionString(ctx) + "," + natsContainer3.MustConnectionString(ctx)

	nc, err := natsgo.Connect(servers, natsgo.MaxReconnects(5), natsgo.ReconnectWait(2*time.Second))
	if err != nil {
		log.Printf("connecting to nats container failed:\n\t%v\n", err)
		return
	}

	// Close connection
	defer nc.Close()

	{
		// Simple Publisher
		err = nc.Publish("foo", []byte("Hello World"))
		if err != nil {
			log.Printf("failed to publish message: %s", err)
			return
		}
	}

	{
		// Channel subscriber
		ch := make(chan *natsgo.Msg, 64)
		sub, err := nc.ChanSubscribe("channel", ch)
		if err != nil {
			log.Printf("failed to subscribe to message: %s", err)
			return
		}

		// Request
		err = nc.Publish("channel", []byte("Hello NATS Cluster!"))
		if err != nil {
			log.Printf("failed to publish message: %s", err)
			return
		}

		msg := <-ch
		fmt.Println(string(msg.Data))

		err = sub.Unsubscribe()
		if err != nil {
			log.Printf("failed to unsubscribe: %s", err)
			return
		}

		err = sub.Drain()
		if err != nil {
			log.Printf("failed to drain: %s", err)
			return
		}
	}

	{
		// Responding to a request message
		sub, err := nc.Subscribe("request", func(m *natsgo.Msg) {
			err1 := m.Respond([]byte("answer is 42"))
			if err1 != nil {
				log.Printf("failed to respond to message: %s", err1)
				return
			}
		})
		if err != nil {
			log.Printf("failed to subscribe to message: %s", err)
			return
		}

		// Request
		msg, err := nc.Request("request", []byte("what is the answer?"), 1*time.Second)
		if err != nil {
			log.Printf("failed to send request: %s", err)
			return
		}

		fmt.Println(string(msg.Data))

		err = sub.Unsubscribe()
		if err != nil {
			log.Printf("failed to unsubscribe: %s", err)
			return
		}

		err = sub.Drain()
		if err != nil {
			log.Printf("failed to drain: %s", err)
			return
		}
	}

	// Drain connection (Preferred for responders)
	// Close() not needed if this is called.
	err = nc.Drain()
	if err != nil {
		log.Printf("failed to drain connection: %s", err)
		return
	}

	// Output:
	// Hello NATS Cluster!
	// answer is 42
}
