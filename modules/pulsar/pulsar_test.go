package pulsar_test

import (
	"context"
	"strings"
	"testing"

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
		})
	}
}
