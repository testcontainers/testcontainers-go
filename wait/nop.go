package wait

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go/exec"
)

var (
	_ Strategy        = (*NopStrategy)(nil)
	_ StrategyTimeout = (*NopStrategy)(nil)
)

type NopStrategy struct {
	timeout        *time.Duration
	waitUntilReady func(context.Context, StrategyTarget) error
}

func ForNop(
	waitUntilReady func(context.Context, StrategyTarget) error,
) *NopStrategy {
	return &NopStrategy{
		waitUntilReady: waitUntilReady,
	}
}

func (ws *NopStrategy) Timeout() *time.Duration {
	return ws.timeout
}

func (ws *NopStrategy) WithStartupTimeout(timeout time.Duration) *NopStrategy {
	ws.timeout = &timeout
	return ws
}

func (ws *NopStrategy) WaitUntilReady(ctx context.Context, target StrategyTarget) error {
	return ws.waitUntilReady(ctx, target)
}

// NopStrategyTargetOption is a functional option for customising a NopStrategyTarget.
type NopStrategyTargetOption func(*NopStrategyTarget)

// NopTargetData sets the ReaderCloser to read from data.
func NopTargetData(data string) NopStrategyTargetOption {
	return func(target *NopStrategyTarget) {
		target.ReaderCloser = io.NopCloser(bytes.NewBufferString(data))
	}
}

// NewNopStrategyTarget returns a fully functional NopStrategyTarget which can be
// used in tests and customised using the provided options.
func NewNopStrategyTarget(options ...NopStrategyTargetOption) *NopStrategyTarget {
	s := &NopStrategyTarget{
		ReaderCloser:   io.NopCloser(new(bytes.Buffer)),
		ContainerState: types.ContainerState{Running: true},
	}

	for _, option := range options {
		option(s)
	}

	return s
}

// NopStrategyTarget is a fully functional StrategyTarget which can be used in tests.
type NopStrategyTarget struct {
	ReaderCloser   io.ReadCloser
	ContainerState types.ContainerState
}

func (st NopStrategyTarget) Host(_ context.Context) (string, error) {
	return "", nil
}

func (st NopStrategyTarget) Inspect(_ context.Context) (*types.ContainerJSON, error) {
	return nil, nil
}

// Deprecated: use Inspect instead
func (st NopStrategyTarget) Ports(_ context.Context) (nat.PortMap, error) {
	return nil, nil
}

func (st NopStrategyTarget) MappedPort(_ context.Context, n nat.Port) (nat.Port, error) {
	return n, nil
}

func (st NopStrategyTarget) Logs(_ context.Context) (io.ReadCloser, error) {
	return st.ReaderCloser, nil
}

func (st NopStrategyTarget) Exec(_ context.Context, _ []string, _ ...exec.ProcessOption) (int, io.Reader, error) {
	return 0, nil, nil
}

func (st NopStrategyTarget) State(_ context.Context) (*types.ContainerState, error) {
	return &st.ContainerState, nil
}

func (st NopStrategyTarget) CopyFileFromContainer(context.Context, string) (io.ReadCloser, error) {
	return st.ReaderCloser, nil
}
