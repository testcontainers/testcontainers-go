package wait

import (
	"context"
	"io"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
)

// Implement interface
var _ Strategy = (*HostPortStrategy)(nil)

type HostPortStrategy struct {
	// all WaitStrategies should have a startupTimeout to avoid waiting infinitely
	startupTimeout time.Duration
}

// NewHostPortStrategy constructs a default host port strategy
func NewHostPortStrategy() *HostPortStrategy {
	return &HostPortStrategy{
		startupTimeout: defaultStartupTimeout(),
	}
}

// fluent builders for each property
// since go has neither covariance nor generics, the return type must be the type of the concrete implementation
// this is true for all properties, even the "shared" ones like startupTimeout

// ForListeningPort is a helper similar to those in Wait.java
// https://github.com/testcontainers/testcontainers-java/blob/1d85a3834bd937f80aad3a4cec249c027f31aeb4/core/src/main/java/org/testcontainers/containers/wait/strategy/Wait.java
func ForListeningPort() *HostPortStrategy {
	return NewHostPortStrategy()
}

func (hp *HostPortStrategy) WithStartupTimeout(startupTimeout time.Duration) *HostPortStrategy {
	hp.startupTimeout = startupTimeout
	return hp
}

// WaitUntilReady implements Strategy.WaitUntilReady
func (hp *HostPortStrategy) WaitUntilReady(ctx context.Context, target StrategyTarget) (err error) {
	// limit context to startupTimeout
	ctx, cancelContext := context.WithTimeout(ctx, hp.startupTimeout)
	defer cancelContext()

	ipAddress, err := target.GetIPAddress(ctx)
	if err != nil {
		return
	}

	ports, err := target.LivenessCheckPorts(ctx)
	if err != nil {
		return
	}

	// Bookkeeping for all opened connections
	var closers []io.Closer
	defer func() {
		for _, closer := range closers {
			closer.Close()
		}
	}()

	for port := range ports {
		proto := port.Proto()
		portNumber := port.Int()
		portString := strconv.Itoa(portNumber)

		dialer := net.Dialer{}

		address := net.JoinHostPort(ipAddress, portString)
		for {
			conn, err := dialer.DialContext(ctx, proto, address)
			if err != nil {
				if v, ok := err.(*net.OpError); ok {
					if v2, ok := (v.Err).(*os.SyscallError); ok {
						if v2.Err == syscall.ECONNREFUSED {
							time.Sleep(100 * time.Millisecond)
							continue
						}
					}
				}
				return err
			}
			closers = append(closers, conn)
			break
		}
	}

	return nil
}
