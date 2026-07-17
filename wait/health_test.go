package wait_test

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

// newStateTarget creates a new mockStrategyTarget whose State method always returns the given state.
func newStateTarget(t *testing.T, state *container.State) *mockStrategyTarget {
	t.Helper()
	target := newMockStrategyTarget(t)
	target.EXPECT().State(anyContext).Return(state, nil)
	return target
}

// TestWaitForHealthTimesOutForUnhealthy confirms that an unhealthy container will eventually
// time out.
func TestWaitForHealthTimesOutForUnhealthy(t *testing.T) {
	target := newStateTarget(t, &container.State{
		Running: true,
		Health:  &container.Health{Status: container.Unhealthy},
	})

	wg := wait.NewHealthStrategy().WithStartupTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

// TestWaitForHealthSucceeds ensures that a healthy container always succeeds.
func TestWaitForHealthSucceeds(t *testing.T) {
	target := newStateTarget(t, &container.State{
		Running: true,
		Health:  &container.Health{Status: container.Healthy},
	})

	wg := wait.NewHealthStrategy().WithStartupTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)

	require.NoError(t, err)
}

// TestWaitForHealthWithNil checks that an initial `nil` Health will not cause a panic,
// and if the container eventually becomes healthy, the HealthStrategy will succeed.
func TestWaitForHealthWithNil(t *testing.T) {
	target := newMockStrategyTarget(t)

	startTime := time.Now()
	target.EXPECT().State(anyContext).RunAndReturn(func(_ context.Context) (*container.State, error) {
		if time.Since(startTime) >= 200*time.Millisecond {
			return &container.State{Running: true, Health: &container.Health{Status: container.Healthy}}, nil
		}
		return &container.State{Running: true, Health: nil}, nil
	})

	wg := wait.NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

// TestWaitFailsForNilHealth checks that Health always nil fails (but will NOT cause a panic)
func TestWaitFailsForNilHealth(t *testing.T) {
	target := newStateTarget(t, &container.State{
		Running: true,
		Health:  nil,
	})

	wg := wait.NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestWaitForHealthFailsDueToOOMKilledContainer(t *testing.T) {
	target := newStateTarget(t, &container.State{
		OOMKilled: true,
	})

	wg := wait.NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.EqualError(t, err, "container crashed with out-of-memory (OOMKilled)")
}

func TestWaitForHealthFailsDueToExitedContainer(t *testing.T) {
	target := newStateTarget(t, &container.State{
		Status:   container.StateExited,
		ExitCode: 1,
	})

	wg := wait.NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.EqualError(t, err, "container exited with code 1")
}

func TestWaitForHealthFailsDueToUnexpectedContainerStatus(t *testing.T) {
	target := newStateTarget(t, &container.State{
		Status: container.StateDead,
	})

	wg := wait.NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.EqualError(t, err, "unexpected container status \"dead\"")
}
