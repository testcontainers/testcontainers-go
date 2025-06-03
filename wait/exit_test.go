package wait

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

type exitStrategyTarget struct {
	isRunning bool
}

func (st exitStrategyTarget) Host(_ context.Context) (string, error) {
	return "", nil
}

func (st exitStrategyTarget) Inspect(_ context.Context) (*container.InspectResponse, error) {
	return nil, nil
}

// Deprecated: use Inspect instead
func (st exitStrategyTarget) Ports(_ context.Context) (nat.PortMap, error) {
	return nil, nil
}

func (st exitStrategyTarget) MappedPort(_ context.Context, n nat.Port) (nat.Port, error) {
	return n, nil
}

func (st exitStrategyTarget) Logs(_ context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func (st exitStrategyTarget) Exec(_ context.Context, _ []string, _ ...tcexec.ProcessOption) (int, io.Reader, error) {
	return 0, nil, nil
}

func (st exitStrategyTarget) State(_ context.Context) (*container.State, error) {
	return &container.State{Running: st.isRunning}, nil
}

func (st exitStrategyTarget) CopyFileFromContainer(context.Context, string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func TestWaitForExit(t *testing.T) {
	target := exitStrategyTarget{
		isRunning: false,
	}
	wg := NewExitStrategy().WithExitTimeout(100 * time.Millisecond)
	err := wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}
