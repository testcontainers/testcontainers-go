package wait

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

type healthStrategyTarget struct {
	state *container.State
	mtx   sync.Mutex
}

func (st *healthStrategyTarget) Host(_ context.Context) (string, error) {
	return "", nil
}

func (st *healthStrategyTarget) Inspect(_ context.Context) (*container.InspectResponse, error) {
	return nil, nil
}

// Deprecated: use Inspect instead
func (st *healthStrategyTarget) Ports(_ context.Context) (nat.PortMap, error) {
	return nil, nil
}

func (st *healthStrategyTarget) MappedPort(_ context.Context, n nat.Port) (nat.Port, error) {
	return n, nil
}

func (st *healthStrategyTarget) Logs(_ context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func (st *healthStrategyTarget) Exec(_ context.Context, _ []string, _ ...tcexec.ProcessOption) (int, io.Reader, error) {
	return 0, nil, nil
}

func (st *healthStrategyTarget) State(_ context.Context) (*container.State, error) {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	// Return a copy to prevent data race.
	state := *st.state

	return &state, nil
}

func (st *healthStrategyTarget) setState(health *container.Health) {
	st.mtx.Lock()
	defer st.mtx.Unlock()
	st.state.Health = health
}

func (st *healthStrategyTarget) CopyFileFromContainer(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

// TestWaitForHealthTimesOutForUnhealthy confirms that an unhealthy container will eventually
// time out.
func TestWaitForHealthTimesOutForUnhealthy(t *testing.T) {
	target := &healthStrategyTarget{
		state: &container.State{
			Running: true,
			Health:  &container.Health{Status: types.Unhealthy},
		},
	}
	wg := NewHealthStrategy().WithStartupTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

// TestWaitForHealthSucceeds ensures that a healthy container always succeeds.
func TestWaitForHealthSucceeds(t *testing.T) {
	target := &healthStrategyTarget{
		state: &container.State{
			Running: true,
			Health:  &container.Health{Status: types.Healthy},
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
		state: &container.State{
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
		target.setState(&container.Health{Status: types.Healthy})
	}(target)

	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

// TestWaitFailsForNilHealth checks that Health always nil fails (but will NOT cause a panic)
func TestWaitFailsForNilHealth(t *testing.T) {
	target := &healthStrategyTarget{
		state: &container.State{
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
		state: &container.State{
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
		state: &container.State{
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
		state: &container.State{
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
