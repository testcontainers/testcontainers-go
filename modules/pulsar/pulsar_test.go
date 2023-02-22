package pulsar

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func TestPulsar(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nwName := "pulsar-test"
	_, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name: nwName,
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name string
		opts []PulsarContainerOptions
	}{
		{
			name: "default",
		},
		{
			name: "with modifiers",
			opts: []PulsarContainerOptions{
				WithConfigModifier(func(config *container.Config) {
					config.Env = append(config.Env, "PULSAR_MEM= -Xms512m -Xmx512m -XX:MaxDirectMemorySize=512m")
				}),
				WithHostConfigModifier(func(hostConfig *container.HostConfig) {
					hostConfig.Resources = container.Resources{
						Memory: 1024 * 1024 * 1024,
					}
				}),
				WithEndpointSettingsModifier(func(settings map[string]*network.EndpointSettings) {
					settings[nwName] = &network.EndpointSettings{
						Aliases: []string{"pulsar"},
					}
				}),
			},
		},
		{
			name: "with functions worker",
			opts: []PulsarContainerOptions{
				WithFunctionsWorker(),
			},
		},
		{
			name: "with transactions",
			opts: []PulsarContainerOptions{
				WithTransactions(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := StartContainer(
				ctx,
				tt.opts...,
			)
			if err != nil {
				t.Fatal(err)
			}

			pc, err := pulsar.NewClient(pulsar.ClientOptions{
				URL:               c.URI,
				OperationTimeout:  30 * time.Second,
				ConnectionTimeout: 30 * time.Second,
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { pc.Close() })

			consumer, err := pc.Subscribe(pulsar.ConsumerOptions{
				Topic:            "test-topic",
				SubscriptionName: "pulsar-test",
				Type:             pulsar.Exclusive,
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { consumer.Close() })

			msgChan := make(chan []byte)
			go func() {
				msg, err := consumer.Receive(ctx)
				if err != nil {
					fmt.Println("failed to receive message", err)
					return
				}
				msgChan <- msg.Payload()
				consumer.Ack(msg)
			}()

			producer, err := pc.CreateProducer(pulsar.ProducerOptions{
				Topic: "test-topic",
			})
			if err != nil {
				t.Fatal(err)
			}

			producer.Send(ctx, &pulsar.ProducerMessage{
				Payload: []byte("hello world"),
			})

			ticker := time.NewTicker(1 * time.Minute)
			select {
			case <-ticker.C:
				t.Fatal("did not receive message in time")
			case msg := <-msgChan:
				if string(msg) != "hello world" {
					t.Fatal("received unexpected message bytes")
				}
			}
		})
	}
}
