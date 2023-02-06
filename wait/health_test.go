package wait

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

type healthStrategyTarget struct {
	Health *types.Health
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
	return &types.ContainerState{Health: st.Health}, nil
}

func TestWaitForHealthTimesOutForUnhealthy(t *testing.T) {
	target := healthStrategyTarget{Health: &types.Health{Status: types.Unhealthy}}
	wg := NewHealthStrategy().WithStartupTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)

	if err != nil && err == context.DeadlineExceeded {
		return
	}
	t.Fatal(err)
}

func TestWaitForHealthSucceeds(t *testing.T) {
	target := healthStrategyTarget{Health: &types.Health{Status: types.Healthy}}
	wg := NewHealthStrategy().WithStartupTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)

	if err != nil {
		t.Fatal(err)
	}
}

func TestWaitForHealthWithNil(t *testing.T) {
	target := &healthStrategyTarget{Health: nil}
	wg := NewHealthStrategy().
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	go func(target *healthStrategyTarget) {
		// wait a bit to simulate startup time and give check time to at least
		// try a few times with a nil Health
		time.Sleep(200 * time.Millisecond)
		target.Health = &types.Health{Status: types.Healthy}
	}(target)

	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
}
