package wait

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
)

type Strategy interface {
	WaitUntilReady(context.Context, StrategyTarget) error
}

type StrategyTarget interface {
	Host(context.Context) (string, error)
	MappedPort(context.Context, nat.Port) (nat.Port, error)
	Logs(context.Context) (io.ReadCloser, error)
	Exec(ctx context.Context, cmd []string) (int, error)
	State(context.Context) (*types.ContainerState, error)
}

func defaultTimeout() time.Duration {
	return 60 * time.Second
}

func defaultPollInterval() time.Duration {
	return 100 * time.Millisecond
}
