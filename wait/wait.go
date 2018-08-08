package wait

import (
	"time"
	"github.com/docker/go-connections/nat"
	"context"
)

type WaitStrategy interface {
	WaitUntilReady(ctx context.Context, waitStrategyTarget WaitStrategyTarget) error
}

type WaitStrategyTarget interface {
	GetIPAddress(ctx context.Context) (string, error)
	LivenessCheckPorts(ctx context.Context) (nat.PortSet, error)
}

func defaultStartupTimeout() time.Duration {
	return 60 * time.Second
}
