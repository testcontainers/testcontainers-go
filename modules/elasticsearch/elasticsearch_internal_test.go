package elasticsearch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_setWaitFor(t *testing.T) {

	sampleTimeout := 42 * time.Second
	tests := []struct {
		name     string
		options  *Options
		expected *time.Duration
	}{
		{
			name:     "when_no_startupTimeout_timeout_is_nil",
			options:  &Options{},
			expected: nil,
		},
		{
			name:     "when_no_startupTimeout_is_set_timeout_is_the_same",
			options:  &Options{startupTimeout: sampleTimeout},
			expected: &sampleTimeout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &testcontainers.ContainerRequest{}

			setWaitFor(tt.options, request)

			actual, ok := request.WaitingFor.(*wait.HTTPStrategy)
			require.True(t, ok, "strategy should be HTTPStrategy")
			assert.Equal(t, tt.expected, actual.Timeout())
		})
	}

	t.Run("when_startupTimeout_and_SSL_are_setup_they_work_together", func(t *testing.T) {
		timeout := 12 * time.Millisecond
		options := &Options{startupTimeout: timeout}
		request := &testcontainers.ContainerRequest{
			Image: "docker.elastic.co/elasticsearch/elasticsearch:8.9.0",
		}

		setWaitFor(options, request)

		actual, ok := request.WaitingFor.(*wait.MultiStrategy)
		require.True(t, ok, "strategy should be HTTPStrategy, was %T", request.WaitingFor)
		require.Len(t, actual.Strategies, 2)
		for i := range actual.Strategies {
			t.Logf("strategy: %T", actual.Strategies[i])
		}

		actualFileStrategy, ok := actual.Strategies[0].(*wait.FileStrategy)
		require.True(t, ok, "strategy should be FileStrategy, was: %T", actual.Strategies[0])
		assert.Equal(t, &timeout, actualFileStrategy.Timeout())

		actualHTTPStrategy, ok := actual.Strategies[1].(*wait.HTTPStrategy)
		require.True(t, ok, "strategy should be HTTPStrategy, was: %T", actual.Strategies[1])
		assert.Equal(t, &timeout, actualHTTPStrategy.Timeout())
	})
	t.Run("when_request_already_has_a_strategy_it_is_first", func(t *testing.T) {
		timeout := 12 * time.Second
		options := &Options{
			startupTimeout: timeout,
		}
		request := &testcontainers.ContainerRequest{
			WaitingFor: wait.ForExit(),
		}

		setWaitFor(options, request)

		actual, ok := request.WaitingFor.(*wait.MultiStrategy)
		require.True(t, ok, "strategy should be HTTPStrategy, was %T", request.WaitingFor)
		require.Len(t, actual.Strategies, 2)
		for i := range actual.Strategies {
			t.Logf("strategy: %T", actual.Strategies[i])
		}

		_, ok = actual.Strategies[0].(*wait.ExitStrategy)
		require.True(t, ok, "strategy should be ExitStrategy, was: %T", actual.Strategies[0])

		actualHTTPStrategy, ok := actual.Strategies[1].(*wait.HTTPStrategy)
		require.True(t, ok, "strategy should be HTTPStrategy, was: %T", actual.Strategies[1])
		assert.Equal(t, &timeout, actualHTTPStrategy.Timeout())
	})
}
