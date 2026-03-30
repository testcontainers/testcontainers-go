package wait_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

// TestWaitForHealthTimesOutForUnhealthy confirms that an unhealthy container will eventually
// time out.
func TestWaitForHealthTimesOutForUnhealthy(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("State", mock.Anything).Return(&container.State{
		Running: true,
		Health:  &container.Health{Status: types.Unhealthy},
	}, nil)
	wg := wait.NewHealthStrategy().WithStartupTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

// TestWaitForHealthSucceeds ensures that a healthy container always succeeds.
func TestWaitForHealthSucceeds(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("State", mock.Anything).Return(&container.State{
		Running: true,
		Health:  &container.Health{Status: types.Healthy},
	}, nil)
	wg := wait.NewHealthStrategy().WithStartupTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)

	require.NoError(t, err)
}

// TestWaitForHealthWithNil checks that an initial `nil` Health will not cause a panic,
// and if the container eventually becomes healthy, the HealthStrategy will succeed.
func TestWaitForHealthWithNil(t *testing.T) {
	var mtx sync.Mutex
	state := &container.State{Running: true, Health: nil}

	target := newMockStrategyTarget(t)
	target.On("State", mock.Anything).RunAndReturn(
		func(_ context.Context) (*container.State, error) {
			mtx.Lock()
			defer mtx.Unlock()
			// Return a copy to prevent data race.
			s := *state
			return &s, nil
		},
	)

	wg := wait.NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	go func() {
		// wait a bit to simulate startup time and give check time to at least
		// try a few times with a nil Health
		time.Sleep(200 * time.Millisecond)
		mtx.Lock()
		state.Health = &container.Health{Status: types.Healthy}
		mtx.Unlock()
	}()

	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

// TestWaitFailsForNilHealth checks that Health always nil fails (but will NOT cause a panic)
func TestWaitFailsForNilHealth(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("State", mock.Anything).Return(&container.State{
		Running: true,
		Health:  nil,
	}, nil)
	wg := wait.NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestWaitForHealthFailsDueToOOMKilledContainer(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("State", mock.Anything).Return(&container.State{
		OOMKilled: true,
	}, nil)
	wg := wait.NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.EqualError(t, err, "container crashed with out-of-memory (OOMKilled)")
}

func TestWaitForHealthFailsDueToExitedContainer(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("State", mock.Anything).Return(&container.State{
		Status:   "exited",
		ExitCode: 1,
	}, nil)
	wg := wait.NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.EqualError(t, err, "container exited with code 1")
}

func TestWaitForHealthFailsDueToUnexpectedContainerStatus(t *testing.T) {
	target := newMockStrategyTarget(t)
	target.On("State", mock.Anything).Return(&container.State{
		Status: "dead",
	}, nil)
	wg := wait.NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.EqualError(t, err, "unexpected container status \"dead\"")
}
