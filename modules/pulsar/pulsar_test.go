package pulsar_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	testcontainerspulsar "github.com/testcontainers/testcontainers-go/modules/pulsar"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
)

// logConsumerForTesting {
// logConsumer is a testcontainers.LogConsumer that prints the log to stdout
type testLogConsumer struct{}

// Accept prints the log to stdout
func (lc *testLogConsumer) Accept(l testcontainers.Log) {
	fmt.Print(string(l.Content))
}

// }

func TestPulsar(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nw, err := tcnetwork.New(ctx, tcnetwork.WithCheckDuplicate())
	require.NoError(t, err)

	nwName := nw.Name

	tests := []struct {
		name         string
		opts         []testcontainers.ContainerCustomizer
		logConsumers []testcontainers.LogConsumer
	}{
		{
			name: "default",
		},
		{
			name: "with modifiers",
			opts: []testcontainers.ContainerCustomizer{
				testcontainers.WithImage("docker.io/apachepulsar/pulsar:2.10.2"),
				// addPulsarEnv {
				testcontainerspulsar.WithPulsarEnv("brokerDeduplicationEnabled", "true"),
				// }
				// advancedDockerSettings {
				testcontainers.WithConfigModifier(func(config *container.Config) {
					config.Env = append(config.Env, "PULSAR_MEM= -Xms512m -Xmx512m -XX:MaxDirectMemorySize=512m")
				}),
				testcontainers.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
					hostConfig.Resources = container.Resources{
						Memory: 1024 * 1024 * 1024,
					}
				}),
				testcontainers.WithEndpointSettingsModifier(func(settings map[string]*network.EndpointSettings) {
					settings[nwName] = &network.EndpointSettings{
						Aliases: []string{"pulsar"},
					}
				}),
				// }
			},
		},
		{
			name: "with functions worker",
			opts: []testcontainers.ContainerCustomizer{
				// withFunctionsWorker {
				testcontainerspulsar.WithFunctionsWorker(),
				// }
			},
		},
		{
			name: "with transactions",
			opts: []testcontainers.ContainerCustomizer{
				// withTransactions {
				testcontainerspulsar.WithTransactions(),
				// }
			},
		},
		{
			name:         "with log consumers",
			logConsumers: []testcontainers.LogConsumer{&testLogConsumer{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := testcontainerspulsar.RunContainer(
				ctx,
				tt.opts...,
			)
			require.NoError(t, err)
			defer func() {
				err := c.Terminate(ctx)
				require.NoError(t, err)
			}()

			// withLogConsumers {
			if len(c.LogConsumers) > 0 {
				c.WithLogConsumers(ctx, tt.logConsumers...)
				defer func() {
					// not handling the error because it will never return an error: it's satisfying the current API
					_ = c.StopLogProducer()
				}()
			}
			// }

			// getBrokerURL {
			brokerURL, err := c.BrokerURL(ctx)
			// }
			require.NoError(t, err)

			// getAdminURL {
			serviceURL, err := c.HTTPServiceURL(ctx)
			// }
			require.NoError(t, err)

			assert.True(t, strings.HasPrefix(brokerURL, "pulsar://"))
			assert.True(t, strings.HasPrefix(serviceURL, "http://"))

			pc, err := pulsar.NewClient(pulsar.ClientOptions{
				URL:               brokerURL,
				OperationTimeout:  30 * time.Second,
				ConnectionTimeout: 30 * time.Second,
			})
			require.NoError(t, err)
			t.Cleanup(func() { pc.Close() })

			subscriptionName := "pulsar-test"

			consumer, err := pc.Subscribe(pulsar.ConsumerOptions{
				Topic:            "test-topic",
				SubscriptionName: subscriptionName,
				Type:             pulsar.Exclusive,
			})
			require.NoError(t, err)
			t.Cleanup(func() { consumer.Close() })

			msgChan := make(chan []byte)
			go func() {
				msg, err := consumer.Receive(ctx)
				if err != nil {
					fmt.Println("failed to receive message", err)
					return
				}
				msgChan <- msg.Payload()
				err = consumer.Ack(msg)
				if err != nil {
					fmt.Println("failed to send ack", err)
					return
				}
			}()

			producer, err := pc.CreateProducer(pulsar.ProducerOptions{
				Topic: "test-topic",
			})
			require.NoError(t, err)

			_, err = producer.Send(ctx, &pulsar.ProducerMessage{
				Payload: []byte("hello world"),
			})
			require.NoError(t, err)

			ticker := time.NewTicker(1 * time.Minute)
			select {
			case <-ticker.C:
				t.Fatal("did not receive message in time")
			case msg := <-msgChan:
				if string(msg) != "hello world" {
					t.Fatal("received unexpected message bytes")
				}
			}

			// get topic statistics using the Admin endpoint
			httpClient := http.Client{
				Timeout: 30 * time.Second,
			}

			resp, err := httpClient.Get(serviceURL + "/admin/v2/persistent/public/default/test-topic/stats")
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var stats map[string]interface{}
			err = json.Unmarshal(body, &stats)
			require.NoError(t, err)

			subscriptions := stats["subscriptions"]
			require.NotNil(t, subscriptions)

			subscriptionsMap := subscriptions.(map[string]interface{})

			// check that the subscription exists
			_, ok := subscriptionsMap[subscriptionName]
			assert.True(t, ok)
		})
	}

	// remove the network after the last, so that all containers are already removed
	// and there are no active endpoints on the network
	t.Cleanup(func() {
		err := nw.Remove(context.Background())
		require.NoError(t, err)
	})
}
