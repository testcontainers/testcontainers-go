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

type exitStrategyTarget struct {
	isRunning bool
}

func (st exitStrategyTarget) Host(ctx context.Context) (string, error) {
	return "", nil
}

func (st exitStrategyTarget) Ports(ctx context.Context) (nat.PortMap, error) {
	return nil, nil
}

func (st exitStrategyTarget) MappedPort(ctx context.Context, n nat.Port) (nat.Port, error) {
	return n, nil
}

func (st exitStrategyTarget) Logs(ctx context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func (st exitStrategyTarget) Exec(ctx context.Context, cmd []string, options ...tcexec.ProcessOption) (int, io.Reader, error) {
	return 0, nil, nil
}

func (st exitStrategyTarget) State(ctx context.Context) (*types.ContainerState, error) {
	return &types.ContainerState{Running: st.isRunning}, nil
}

func TestWaitForExit(t *testing.T) {
	target := exitStrategyTarget{
		isRunning: false,
	}
	wg := NewExitStrategy().WithExitTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
}
