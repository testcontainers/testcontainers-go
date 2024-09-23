package consul_test

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	capi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/consul"
)

func TestConsul(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name string
		opts []testcontainers.ContainerCustomizer
	}{
		{
			name: "Default",
			opts: []testcontainers.ContainerCustomizer{},
		},
		{
			name: "WithConfigString",
			opts: []testcontainers.ContainerCustomizer{
				consul.WithConfigString(`{ "server":true }`),
			},
		},
		{
			name: "WithConfigFile",
			opts: []testcontainers.ContainerCustomizer{
				consul.WithConfigFile(filepath.Join("testdata", "config.json")),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctr, err := consul.Run(ctx, "docker.io/hashicorp/consul:1.15", test.opts...)
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			// Check if API is up
			host, err := ctr.ApiEndpoint(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, host)

			res, err := http.Get("http://" + host)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			cfg := capi.DefaultConfig()
			cfg.Address = host

			reg := &capi.AgentServiceRegistration{ID: "abcd", Name: test.name}

			client, err := capi.NewClient(cfg)
			require.NoError(t, err)

			// Register / Unregister service
			s := client.Agent()
			err = s.ServiceRegister(reg)
			require.NoError(t, err)

			err = s.ServiceDeregister("abcd")
			require.NoError(t, err)
		})
	}
}
