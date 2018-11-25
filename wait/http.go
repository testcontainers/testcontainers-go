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
var _ Strategy = (*HTTPStrategy)(nil)

type HTTPStrategy struct {
	// all Strategies should have a startupTimeout to avoid waiting infinitely
	startupTimeout time.Duration

	// additional properties
	Path              string
	StatusCodeMatcher func(status int) bool
	UseTLS            bool
}

// NewHTTPStrategy constructs a HTTP strategy waiting on port 80 and status code 200
func NewHTTPStrategy(path string) *HTTPStrategy {
	return &HTTPStrategy{
		startupTimeout:    defaultStartupTimeout(),
		Path:              path,
		StatusCodeMatcher: defaultStatusCodeMatcher,
		UseTLS:            false,
	}

}

func defaultStatusCodeMatcher(status int) bool {
	return status == http.StatusOK
}

// fluent builders for each property
// since go has neither covariance nor generics, the return type must be the type of the concrete implementation
// this is true for all properties, even the "shared" ones like startupTimeout

func (ws *HTTPStrategy) WithStartupTimeout(startupTimeout time.Duration) *HTTPStrategy {
	ws.startupTimeout = startupTimeout
	return ws
}

func (ws *HTTPStrategy) WithStatusCodeMatcher(statusCodeMatcher func(status int) bool) *HTTPStrategy {
	ws.StatusCodeMatcher = statusCodeMatcher
	return ws
}

func (ws *HTTPStrategy) WithTLS(useTLS bool) *HTTPStrategy {
	ws.UseTLS = useTLS
	return ws
}

// ForHTTP is a convenience method similar to Wait.java
// https://github.com/testcontainers/testcontainers-java/blob/1d85a3834bd937f80aad3a4cec249c027f31aeb4/core/src/main/java/org/testcontainers/containers/wait/strategy/Wait.java
func ForHTTP(path string) *HTTPStrategy {
	return NewHTTPStrategy(path)
}

// WaitUntilReady implements Strategy.WaitUntilReady
func (ws *HTTPStrategy) WaitUntilReady(ctx context.Context, target StrategyTarget) (err error) {
	// limit context to startupTimeout
	ctx, cancelContext := context.WithTimeout(ctx, ws.startupTimeout)
	defer cancelContext()

	ipAddress, err := target.Host(ctx)
	if err != nil {
		return
	}

	ports, err := target.Ports(ctx)
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
	if ws.UseTLS {
		proto = "https"
	} else {
		proto = "http"
	}

	url := fmt.Sprintf("%s://%s%s", proto, potentialAddresses[0], ws.Path)

	client := http.Client{Timeout: ws.startupTimeout}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
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

		if !ws.StatusCodeMatcher(resp.StatusCode) {
			continue
		}

		break
	}

	return nil
}
