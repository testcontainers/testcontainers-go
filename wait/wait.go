package wait

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go/exec"
)

// Strategy defines the basic interface for a Wait Strategy
type Strategy interface {
	WaitUntilReady(context.Context, StrategyTarget) error
}

// StrategyTimeout allows MultiStrategy to configure a Strategy's Timeout
type StrategyTimeout interface {
	Timeout() *time.Duration
}

type StrategyTarget interface {
	Host(context.Context) (string, error)
	Ports(ctx context.Context) (nat.PortMap, error)
	MappedPort(context.Context, nat.Port) (nat.Port, error)
	Logs(context.Context) (io.ReadCloser, error)
	Exec(context.Context, []string, ...exec.ProcessOption) (int, io.Reader, error)
	State(context.Context) (*types.ContainerState, error)
}

func defaultStartupTimeout() time.Duration {
	return 60 * time.Second
}

func defaultPollInterval() time.Duration {
	return 100 * time.Millisecond
}
