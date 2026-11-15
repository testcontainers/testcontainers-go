package solace_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	sc "github.com/testcontainers/testcontainers-go/modules/solace"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRunWaitStrategyDeadline(t *testing.T) {
	t.Run("default-services", func(t *testing.T) {
		testRunWaitStrategyDeadline(t, 4*time.Minute)
	})

	t.Run("single-service", func(t *testing.T) {
		testRunWaitStrategyDeadline(t, 150*time.Second, sc.WithServices(sc.ServiceAMQP))
	})

	t.Run("multiple-services", func(t *testing.T) {
		testRunWaitStrategyDeadline(t, 210*time.Second, sc.WithServices(sc.ServiceAMQP, sc.ServiceManagement, sc.ServiceSMF))
	})

	t.Run("custom-deadline", func(t *testing.T) {
		testRunWaitStrategyDeadline(t, 5*time.Minute,
			testcontainers.WithWaitStrategyAndDeadline(5*time.Minute, wait.ForLog("ready")),
		)
	})
}

func testRunWaitStrategyDeadline(t *testing.T, timeout time.Duration, opts ...testcontainers.ContainerCustomizer) {
	t.Helper()

	stop := errors.New("stop before container creation")
	opts = append(opts, testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
		strategy, ok := req.WaitingFor.(*wait.MultiStrategy)
		require.True(t, ok)

		// Observe the composed deadline without starting Docker or waiting for Solace.
		before := time.Now()
		strategy.Strategies = []wait.Strategy{wait.ForNop(func(ctx context.Context, _ wait.StrategyTarget) error {
			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			require.WithinRange(t, deadline, before.Add(timeout), time.Now().Add(timeout))
			return nil
		})}
		require.NoError(t, strategy.WaitUntilReady(t.Context(), nil))
		return stop
	}))

	ctr, err := sc.Run(t.Context(), "solace/solace-pubsub-standard:latest", opts...)
	require.ErrorIs(t, err, stop)
	require.Nil(t, ctr)
}
