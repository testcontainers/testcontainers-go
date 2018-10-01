package wait

import (
	"time"
	"net"
	"context"
	"strconv"
	"io"
	"os"
	"syscall"
)

// Implement interface
var _ WaitStrategy = (*hostPortWaitStrategy)(nil)

type hostPortWaitStrategy struct {
	// all WaitStrategies should have a startupTimeout to avoid waiting infinitely
	startupTimeout time.Duration
}

// Constructor
func HostPortWaitStrategyNew() *hostPortWaitStrategy {
	return &hostPortWaitStrategy{
		startupTimeout: defaultStartupTimeout(),
	}
}

// fluent builders for each property
// since go has neither covariance nor generics, the return type must be the type of the concrete implementation
// this is true for all properties, even the "shared" ones like startupTimeout
func (hp *hostPortWaitStrategy) WithStartupTimeout(startupTimeout time.Duration) *hostPortWaitStrategy {
	hp.startupTimeout = startupTimeout
	return hp
}

// Convenience method similar to Wait.java
// https://github.com/testcontainers/testcontainers-java/blob/1d85a3834bd937f80aad3a4cec249c027f31aeb4/core/src/main/java/org/testcontainers/containers/wait/strategy/Wait.java
func ForListeningPort() *hostPortWaitStrategy {
	return HostPortWaitStrategyNew()
}

// Implementation of WaitStrategy.WaitUntilReady
func (hp *hostPortWaitStrategy) WaitUntilReady(ctx context.Context, waitStrategyTarget WaitStrategyTarget) (err error) {
	// limit context to startupTimeout
	ctx, _ = context.WithTimeout(ctx, hp.startupTimeout)

	ipAddress, err := waitStrategyTarget.GetIPAddress(ctx)
	if err != nil {
		return
	}

	ports, err := waitStrategyTarget.LivenessCheckPorts(ctx)
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
