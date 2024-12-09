package kafka

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestConfigureQuorumVoters(t *testing.T) {
	testConfigureControllerQuorumVotersFn := func(t *testing.T, req *testcontainers.GenericContainerRequest, expectedVoters string) {
		t.Helper()

		configureControllerQuorumVoters(req)
		require.Equalf(t, expectedVoters, req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"], "expected KAFKA_CONTROLLER_QUORUM_VOTERS to be %s, got %s", expectedVoters, req.Env["KAFKA_CONTROLLER_QUORUM_VOTERS"])
	}

	t.Run("voters/localhost", func(t *testing.T) {
		testConfigureControllerQuorumVotersFn(t, &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Env: map[string]string{},
			},
		}, "1@localhost:9094")
	})

	t.Run("voters/first-network-alias/first-network", func(t *testing.T) {
		testConfigureControllerQuorumVotersFn(t, &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Env:      map[string]string{},
				Networks: []string{"foo", "bar", "baaz"},
				NetworkAliases: map[string][]string{
					"foo":  {"foo0", "foo1", "foo2", "foo3"},
					"bar":  {"bar0", "bar1", "bar2", "bar3"},
					"baaz": {"baaz0", "baaz1", "baaz2", "baaz3"},
				},
			},
		}, "1@foo0:9094")
	})

	t.Run("voters/localhost/alias-no-networks", func(t *testing.T) {
		testConfigureControllerQuorumVotersFn(t, &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				NetworkAliases: map[string][]string{
					"foo":  {"foo0", "foo1", "foo2", "foo3"},
					"bar":  {"bar0", "bar1", "bar2", "bar3"},
					"baaz": {"baaz0", "baaz1", "baaz2", "baaz3"},
				},
			},
		}, "1@localhost:9094")
	})
}

func TestValidateKRaftVersion(t *testing.T) {
	t.Run("official/valid-version", func(t *testing.T) {
		err := validateKRaftVersion("confluentinc/confluent-local:7.5.0")
		require.NoError(t, err)
	})

	t.Run("official/valid-limit-version", func(t *testing.T) {
		err := validateKRaftVersion("confluentinc/confluent-local:7.4.0")
		require.NoError(t, err)
	})

	t.Run("official/invalid-low-version", func(t *testing.T) {
		err := validateKRaftVersion("confluentinc/confluent-local:7.3.99")
		require.Error(t, err)
	})

	t.Run("official/invalid-too-low-version", func(t *testing.T) {
		err := validateKRaftVersion("confluentinc/confluent-local:5.0.0")
		require.Error(t, err)
	})

	t.Run("unofficial/does-not-validate-KRaft-version", func(t *testing.T) {
		err := validateKRaftVersion("my-kafka:1.0.0")
		require.NoError(t, err)
	})
}

func TestValidateListeners(t *testing.T) {
	t.Run("fail/reserved-listener/port-9093", func(t *testing.T) {
		err := validateListeners(Listener{
			Name: "PLAINTEXT",
			Host: "kafka",
			Port: "9093",
		})
		require.Error(t, err)
	})

	t.Run("fail/reserved-listener/port-9094", func(t *testing.T) {
		err := validateListeners(Listener{
			Name: "PLAINTEXT",
			Host: "kafka",
			Port: "9094",
		})
		require.Error(t, err)
	})

	t.Run("fail/reserved-listener/name-controller", func(t *testing.T) {
		err := validateListeners(Listener{
			Name: "  cOnTrOller   ",
			Host: "kafka",
			Port: "9092",
		})
		require.Error(t, err)
	})

	t.Run("fail/reserved-listener/name-plaintext", func(t *testing.T) {
		err := validateListeners(Listener{
			Name: "plaintext",
			Host: "kafka",
			Port: "9092",
		})
		require.Error(t, err)
	})

	t.Run("fail/port-duplication", func(t *testing.T) {
		err := validateListeners(Listener{
			Name: "test",
			Host: "kafka",
			Port: "9092",
		}, Listener{
			Name: "test2",
			Host: "kafka",
			Port: "9092",
		})
		require.Error(t, err)
	})

	t.Run("fail/name-duplication", func(t *testing.T) {
		err := validateListeners(Listener{
			Name: "test",
			Host: "kafka",
			Port: "9092",
		}, Listener{
			Name: "test",
			Host: "kafka",
			Port: "9095",
		})
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		err := validateListeners(Listener{
			Name: "test",
			Host: "kafka",
			Port: "9092",
		})
		require.NoError(t, err)
	})
}
