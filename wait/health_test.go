package wait

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

type healthStrategyTarget struct {
	state *types.ContainerState
}

func (st healthStrategyTarget) Host(ctx context.Context) (string, error) {
	return "", nil
}

func (st healthStrategyTarget) Ports(ctx context.Context) (nat.PortMap, error) {
	return nil, nil
}

func (st healthStrategyTarget) MappedPort(ctx context.Context, n nat.Port) (nat.Port, error) {
	return n, nil
}

func (st healthStrategyTarget) Logs(ctx context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func (st healthStrategyTarget) Exec(ctx context.Context, cmd []string, options ...tcexec.ProcessOption) (int, io.Reader, error) {
	return 0, nil, nil
}

func (st healthStrategyTarget) State(ctx context.Context) (*types.ContainerState, error) {
	return st.state, nil
}

// TestWaitForHealthTimesOutForUnhealthy confirms that an unhealthy container will eventually
// time out.
func TestWaitForHealthTimesOutForUnhealthy(t *testing.T) {
	target := healthStrategyTarget{
		state: &types.ContainerState{
			Running: true,
			Health:  &types.Health{Status: types.Unhealthy},
		},
	}
	wg := NewHealthStrategy().WithStartupTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

// TestWaitForHealthSucceeds ensures that a healthy container always succeeds.
func TestWaitForHealthSucceeds(t *testing.T) {
	target := healthStrategyTarget{
		state: &types.ContainerState{
			Running: true,
			Health:  &types.Health{Status: types.Healthy},
		},
	}
	wg := NewHealthStrategy().WithStartupTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)

	require.NoError(t, err)
}

// TestWaitForHealthWithNil checks that an initial `nil` Health will not cause a panic,
// and if the container eventually becomes healthy, the HealthStrategy will succeed.
func TestWaitForHealthWithNil(t *testing.T) {
	target := &healthStrategyTarget{
		state: &types.ContainerState{
			Running: true,
			Health:  nil,
		},
	}
	wg := NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	go func(target *healthStrategyTarget) {
		// wait a bit to simulate startup time and give check time to at least
		// try a few times with a nil Health
		time.Sleep(200 * time.Millisecond)
		target.state.Health = &types.Health{Status: types.Healthy}
	}(target)

	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

// TestWaitFailsForNilHealth checks that Health always nil fails (but will NOT cause a panic)
func TestWaitFailsForNilHealth(t *testing.T) {
	target := &healthStrategyTarget{
		state: &types.ContainerState{
			Running: true,
			Health:  nil,
		},
	}
	wg := NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestWaitForHealthFailsDueToOOMKilledContainer(t *testing.T) {
	target := &healthStrategyTarget{
		state: &types.ContainerState{
			OOMKilled: true,
		},
	}
	wg := NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.EqualError(t, err, "container crashed with out-of-memory (OOMKilled)")
}

func TestWaitForHealthFailsDueToExitedContainer(t *testing.T) {
	target := &healthStrategyTarget{
		state: &types.ContainerState{
			Status:   "exited",
			ExitCode: 1,
		},
	}
	wg := NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.EqualError(t, err, "container exited with code 1")
}

func TestWaitForHealthFailsDueToUnexpectedContainerStatus(t *testing.T) {
	target := &healthStrategyTarget{
		state: &types.ContainerState{
			Status: "dead",
		},
	}
	wg := NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	err := wg.WaitUntilReady(context.Background(), target)
	require.Error(t, err)
	require.EqualError(t, err, "unexpected container status \"dead\"")
}
