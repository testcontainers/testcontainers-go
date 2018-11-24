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
	GetIPAddress(ctx context.Context) (string, error)
	LivenessCheckPorts(ctx context.Context) (nat.PortSet, error)
}

func defaultStartupTimeout() time.Duration {
	return 60 * time.Second
}
