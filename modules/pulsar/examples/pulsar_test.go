package examples_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	testcontainerspulsar "github.com/testcontainers/testcontainers-go/modules/pulsar"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
)

// noopLogConsumer implements testcontainers.LogConsumer
// and does nothing with the logs.
type noopLogConsumer struct{}

// Accept implements testcontainers.LogConsumer.
func (*noopLogConsumer) Accept(testcontainers.Log) {}

func TestPulsar(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nw, err := tcnetwork.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, nw)

	nwName := nw.Name

	tests := []struct {
		name string
		opts []testcontainers.ContainerCustomizer
	}{
		{
			name: "default",
		},
		{
			name: "with modifiers",
			opts: []testcontainers.ContainerCustomizer{
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
			name: "with log consumers",
			opts: []testcontainers.ContainerCustomizer{
				// withLogConsumers {
				testcontainers.WithLogConsumers(&noopLogConsumer{}),
				// }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := testcontainerspulsar.Run(
				ctx,
				"apachepulsar/pulsar:4.0.9",
				tt.opts...,
			)
			testcontainers.CleanupContainer(t, c)
			require.NoError(t, err)

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
				Logger:            log.DefaultNopLogger(),
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

			receiveCtx, receiveCancel := context.WithTimeout(ctx, time.Minute)
			type delivery struct {
				payload []byte
				err     error
				ackErr  error
			}
			msgChan := make(chan delivery, 1)
			receiveDone := make(chan struct{})
			go func() {
				defer close(receiveDone)
				msg, err := consumer.Receive(receiveCtx)
				if err != nil {
					msgChan <- delivery{err: err}
					return
				}
				msgChan <- delivery{payload: msg.Payload(), ackErr: consumer.Ack(msg)}
			}()
			t.Cleanup(func() {
				receiveCancel()
				<-receiveDone
			})

			producer, err := pc.CreateProducer(pulsar.ProducerOptions{
				Topic: "test-topic",
			})
			require.NoError(t, err)

			t.Cleanup(producer.Close)

			_, err = producer.Send(receiveCtx, &pulsar.ProducerMessage{
				Payload: []byte("hello world"),
			})
			require.NoError(t, err)

			select {
			case <-receiveCtx.Done():
				t.Fatal("did not receive message in time")
			case msg := <-msgChan:
				require.NoError(t, msg.err)
				require.NoError(t, msg.ackErr)
				require.Equalf(t, "hello world", string(msg.payload), "received unexpected message bytes")
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

			var stats map[string]any
			err = json.Unmarshal(body, &stats)
			require.NoError(t, err)

			subscriptions := stats["subscriptions"]
			require.NotNil(t, subscriptions)

			subscriptionsMap := subscriptions.(map[string]any)

			// check that the subscription exists
			_, ok := subscriptionsMap[subscriptionName]
			require.True(t, ok)
		})
	}
}
