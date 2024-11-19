package kafka

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

func TestValidateKRaftVersion(t *testing.T) {
	tests := []struct {
		name    string
		image   string
		wantErr bool
	}{
		{
			name:    "Official: valid version",
			image:   "confluentinc/confluent-local:7.5.0",
			wantErr: false,
		},
		{
			name:    "Official: valid, limit version",
			image:   "confluentinc/confluent-local:7.4.0",
			wantErr: false,
		},
		{
			name:    "Official: invalid, low version",
			image:   "confluentinc/confluent-local:7.3.99",
			wantErr: true,
		},
		{
			name:    "Official: invalid, too low version",
			image:   "confluentinc/confluent-local:5.0.0",
			wantErr: true,
		},
		{
			name:    "Unofficial does not validate KRaft version",
			image:   "my-kafka:1.0.0",
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateKRaftVersion(test.image)

			if test.wantErr {
				require.Errorf(t, err, "expected error, got nil")
			} else {
				require.NoErrorf(t, err, "expected no error, got %s", err)
			}
		})
	}
}
