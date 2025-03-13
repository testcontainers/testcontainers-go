package wait

import (
	"context"
	"errors"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

var ErrPortNotFound = errors.New("port not found")

type MockStrategyTarget struct {
	HostImpl                  func(context.Context) (string, error)
	InspectImpl               func(context.Context) (*container.InspectResponse, error)
	PortsImpl                 func(context.Context) (nat.PortMap, error)
	MappedPortImpl            func(context.Context, nat.Port) (nat.Port, error)
	LogsImpl                  func(context.Context) (io.ReadCloser, error)
	ExecImpl                  func(context.Context, []string, ...tcexec.ProcessOption) (int, io.Reader, error)
	StateImpl                 func(context.Context) (*container.State, error)
	CopyFileFromContainerImpl func(context.Context, string) (io.ReadCloser, error)
}

func (st MockStrategyTarget) Host(ctx context.Context) (string, error) {
	return st.HostImpl(ctx)
}

func (st MockStrategyTarget) Inspect(ctx context.Context) (*container.InspectResponse, error) {
	return st.InspectImpl(ctx)
}

// Deprecated: use Inspect instead
func (st MockStrategyTarget) Ports(ctx context.Context) (nat.PortMap, error) {
	inspect, err := st.InspectImpl(ctx)
	if err != nil {
		return nil, err
	}

	return inspect.NetworkSettings.Ports, nil
}

func (st MockStrategyTarget) MappedPort(ctx context.Context, port nat.Port) (nat.Port, error) {
	return st.MappedPortImpl(ctx, port)
}

func (st MockStrategyTarget) Logs(ctx context.Context) (io.ReadCloser, error) {
	return st.LogsImpl(ctx)
}

func (st MockStrategyTarget) Exec(ctx context.Context, cmd []string, options ...tcexec.ProcessOption) (int, io.Reader, error) {
	return st.ExecImpl(ctx, cmd, options...)
}

func (st MockStrategyTarget) State(ctx context.Context) (*container.State, error) {
	return st.StateImpl(ctx)
}

func (st MockStrategyTarget) CopyFileFromContainer(ctx context.Context, filePath string) (io.ReadCloser, error) {
	return st.CopyFileFromContainerImpl(ctx, filePath)
}
