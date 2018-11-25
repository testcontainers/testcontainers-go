package wait

import (
	"context"
	"time"

	"github.com/docker/go-connections/nat"
)

type Strategy interface {
	WaitUntilReady(context.Context, StrategyTarget) error
}

type StrategyTarget interface {
	Host(context.Context) (string, error)
	Ports(context.Context) (nat.PortMap, error)
}

func defaultStartupTimeout() time.Duration {
	return 60 * time.Second
}
