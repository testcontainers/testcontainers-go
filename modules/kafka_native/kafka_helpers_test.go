package kafka_native

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestConfigureQuorumVoters(t *testing.T) {
	tests := []struct {
		name           string
		req            *testcontainers.GenericContainerRequest
		expectedVoters string
	}{
		{
			name: "voters on localhost",
			req: &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Env: map[string]string{},
				},
			},
			expectedVoters: "1@localhost:9094",
		},
		{
			name: "voters on first network alias of the first network",
			req: &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Env:      map[string]string{},
					Networks: []string{"foo", "bar", "baaz"},
					NetworkAliases: map[string][]string{
						"foo":  {"foo0", "foo1", "foo2", "foo3"},
						"bar":  {"bar0", "bar1", "bar2", "bar3"},
						"baaz": {"baaz0", "baaz1", "baaz2", "baaz3"},
					},
				},
			},
			expectedVoters: "1@foo0:9094",
		},
		{
			name: "voters on localhost if alias but no networks",
			req: &testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					NetworkAliases: map[string][]string{
						"foo":  {"foo0", "foo1", "foo2", "foo3"},
						"bar":  {"bar0", "bar1", "bar2", "bar3"},
						"baaz": {"baaz0", "baaz1", "baaz2", "baaz3"},
					},
				},
			},
			expectedVoters: "1@localhost:9094",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configureControllerQuorumVoters(test.req)

			require.Equalf(t, test.expectedVoters, test.req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"], "expected KAFKA_CONTROLLER_QUORUM_VOTERS to be %s, got %s", test.expectedVoters, test.req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"])
		})
	}
}
