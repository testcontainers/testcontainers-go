package consul_test

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	capi "github.com/hashicorp/consul/api"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

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
			container, err := consul.Run(ctx, "docker.io/hashicorp/consul:1.15", test.opts...)
			assert.NilError(t, err)
			t.Cleanup(func() { assert.NilError(t, container.Terminate(ctx), "failed to terminate container") })

			// Check if API is up
			host, err := container.ApiEndpoint(ctx)
			assert.NilError(t, err)
			assert.Check(t, len(len(host)) != 0)

			res, err := http.Get("http://" + host)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(http.StatusOK, res.StatusCode))

			cfg := capi.DefaultConfig()
			cfg.Address = host

			reg := &capi.AgentServiceRegistration{ID: "abcd", Name: test.name}

			client, err := capi.NewClient(cfg)
			assert.NilError(t, err)

			// Register / Unregister service
			s := client.Agent()
			err = s.ServiceRegister(reg)
			assert.NilError(t, err)

			err = s.ServiceDeregister("abcd")
			assert.NilError(t, err)
		})
	}
}
