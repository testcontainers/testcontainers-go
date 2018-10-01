package wait

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
)

// Implement interface
var _ WaitStrategy = (*httpWaitStrategy)(nil)

type httpWaitStrategy struct {
	// all WaitStrategies should have a startupTimeout to avoid waiting infinitely
	startupTimeout time.Duration

	// httpWaitStrategy has additional properties
	path              string
	statusCodeMatcher func(status int) bool
	useTLS            bool
}

// Constructor
func HttpWaitStrategyNew(path string) *httpWaitStrategy {
	return &httpWaitStrategy{
		startupTimeout:    defaultStartupTimeout(),
		statusCodeMatcher: defaultStatusCodeMatcher,
		path:              path,
		useTLS:            false,
	}

}

func defaultStatusCodeMatcher(status int) bool {
	return status == http.StatusOK
}


// fluent builders for each property
// since go has neither covariance nor generics, the return type must be the type of the concrete implementation
// this is true for all properties, even the "shared" ones like startupTimeout
func (ws *httpWaitStrategy) WithStartupTimeout(startupTimeout time.Duration) *httpWaitStrategy {
	ws.startupTimeout = startupTimeout
	return ws
}

func (ws *httpWaitStrategy) WithStatusCodeMatcher(statusCodeMatcher func(status int) bool) *httpWaitStrategy {
	ws.statusCodeMatcher = statusCodeMatcher
	return ws
}

func (ws *httpWaitStrategy) WithTLS(useTLS bool) *httpWaitStrategy {
	ws.useTLS = useTLS
	return ws
}

// Convenience method similar to Wait.java
// https://github.com/testcontainers/testcontainers-java/blob/1d85a3834bd937f80aad3a4cec249c027f31aeb4/core/src/main/java/org/testcontainers/containers/wait/strategy/Wait.java
func ForHttp(path string) *httpWaitStrategy {
	return HttpWaitStrategyNew(path)
}

// Implementation of WaitStrategy.WaitUntilReady
func (ws *httpWaitStrategy) WaitUntilReady(ctx context.Context, waitStrategyTarget WaitStrategyTarget) (err error) {
	// limit context to startupTimeout
	ctx, _ = context.WithTimeout(ctx, ws.startupTimeout)

	ipAddress, err := waitStrategyTarget.GetIPAddress(ctx)
	if err != nil {
		return
	}

	ports, err := waitStrategyTarget.LivenessCheckPorts(ctx)
	if err != nil {
		return
	}

	potentialAddresses := make([]string, 0)
	for port := range ports {
		proto := port.Proto()
		if proto != "tcp" {
			continue
		}

		portNumber := port.Int()
		portString := strconv.Itoa(portNumber)

		address := net.JoinHostPort(ipAddress, portString)
		potentialAddresses = append(potentialAddresses, address)
	}

	var proto string
	if ws.useTLS {
		proto = "https"
	} else {
		proto = "http"
	}

	url := fmt.Sprintf("%s://%s%s", proto, potentialAddresses[0], ws.path)

	client := http.Client{Timeout: ws.startupTimeout}
	req, e := http.NewRequest("GET", url, nil)
	if e != nil {
		return e
	}

	req = req.WithContext(ctx)

	switch {
	case len(potentialAddresses) == 0:
		return errors.New("No TCP Liveness Check Port available")
	case len(potentialAddresses) > 1:
		return errors.New("More than TCP Liveness Check Port currently not supported by http.httpWaitStrategy")
	}


	for {
		resp, err := client.Do(req)

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

		if !ws.statusCodeMatcher(resp.StatusCode) {
			continue
		}

		break
	}

	return nil
}
