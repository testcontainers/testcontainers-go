package wait

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/docker/go-connections/nat"
)

// Implement interface
var (
	_ Strategy        = (*HostPortStrategy)(nil)
	_ StrategyTimeout = (*HostPortStrategy)(nil)
)

var errShellNotExecutable = errors.New("/bin/sh command not executable")

type HostPortStrategy struct {
	// Port is a string containing port number and protocol in the format "80/tcp"
	// which
	Port nat.Port
	// all WaitStrategies should have a startupTimeout to avoid waiting infinitely
	timeout      *time.Duration
	PollInterval time.Duration

	// skipInternalCheck is a flag to skip the internal check, which is useful when
	// a shell is not available in the container or when the container doesn't bind
	// the port internally until additional conditions are met.
	skipInternalCheck bool
}

// NewHostPortStrategy constructs a default host port strategy that waits for the given
// port to be exposed. The default startup timeout is 60 seconds.
func NewHostPortStrategy(port nat.Port) *HostPortStrategy {
	return &HostPortStrategy{
		Port:         port,
		PollInterval: defaultPollInterval(),
	}
}

// fluent builders for each property
// since go has neither covariance nor generics, the return type must be the type of the concrete implementation
// this is true for all properties, even the "shared" ones like startupTimeout

// ForListeningPort returns a host port strategy that waits for the given port
// to be exposed and bound internally the container.
// Alias for `NewHostPortStrategy(port)`.
func ForListeningPort(port nat.Port) *HostPortStrategy {
	return NewHostPortStrategy(port)
}

// ForExposedPort returns a host port strategy that waits for the first port
// to be exposed and bound internally the container.
func ForExposedPort() *HostPortStrategy {
	return NewHostPortStrategy("")
}

// SkipInternalCheck changes the host port strategy to skip the internal check,
// which is useful when a shell is not available in the container or when the
// container doesn't bind the port internally until additional conditions are met.
func (hp *HostPortStrategy) SkipInternalCheck() *HostPortStrategy {
	hp.skipInternalCheck = true

	return hp
}

// WithStartupTimeout can be used to change the default startup timeout
func (hp *HostPortStrategy) WithStartupTimeout(startupTimeout time.Duration) *HostPortStrategy {
	hp.timeout = &startupTimeout
	return hp
}

// WithPollInterval can be used to override the default polling interval of 100 milliseconds
func (hp *HostPortStrategy) WithPollInterval(pollInterval time.Duration) *HostPortStrategy {
	hp.PollInterval = pollInterval
	return hp
}

func (hp *HostPortStrategy) Timeout() *time.Duration {
	return hp.timeout
}

// WaitUntilReady implements Strategy.WaitUntilReady
func (hp *HostPortStrategy) WaitUntilReady(ctx context.Context, target StrategyTarget) error {
	timeout := defaultStartupTimeout()
	if hp.timeout != nil {
		timeout = *hp.timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ipAddress, err := target.Host(ctx)
	if err != nil {
		return err
	}

	waitInterval := hp.PollInterval

	internalPort := hp.Port
	if internalPort == "" {
		inspect, err := target.Inspect(ctx)
		if err != nil {
			return err
		}

		for port := range inspect.NetworkSettings.Ports {
			if internalPort == "" || port.Int() < internalPort.Int() {
				internalPort = port
			}
		}
	}

	if internalPort == "" {
		return fmt.Errorf("no port to wait for")
	}

	var port nat.Port
	port, err = target.MappedPort(ctx, internalPort)
	i := 0

	for port == "" {
		i++

		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", ctx.Err(), err)
		case <-time.After(waitInterval):
			if err := checkTarget(ctx, target); err != nil {
				return err
			}
			port, err = target.MappedPort(ctx, internalPort)
			if err != nil {
				log.Printf("(%d) [%s] %s\n", i, port, err)
			}
		}
	}

	if err := externalCheck(ctx, ipAddress, port, target, waitInterval); err != nil {
		return err
	}

	if hp.skipInternalCheck {
		return nil
	}

	err = internalCheck(ctx, internalPort, target)
	if err != nil && errors.Is(errShellNotExecutable, err) {
		log.Println("Shell not executable in container, only external port check will be performed")
	} else {
		return err
	}

	return nil
}

func externalCheck(ctx context.Context, ipAddress string, port nat.Port, target StrategyTarget, waitInterval time.Duration) error {
	proto := port.Proto()
	portNumber := port.Int()
	portString := strconv.Itoa(portNumber)

	dialer := net.Dialer{}
	address := net.JoinHostPort(ipAddress, portString)
	for {
		if err := checkTarget(ctx, target); err != nil {
			return err
		}
		conn, err := dialer.DialContext(ctx, proto, address)
		if err != nil {
			var v *net.OpError
			if errors.As(err, &v) {
				var v2 *os.SyscallError
				if errors.As(v.Err, &v2) {
					if isConnRefusedErr(v2.Err) {
						time.Sleep(waitInterval)
						continue
					}
				}
			}
			return err
		}

		conn.Close()
		return nil
	}
}

func internalCheck(ctx context.Context, internalPort nat.Port, target StrategyTarget) error {
	command := buildInternalCheckCommand(internalPort.Int())
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := checkTarget(ctx, target); err != nil {
			return err
		}
		exitCode, _, err := target.Exec(ctx, []string{"/bin/sh", "-c", command})
		if err != nil {
			return fmt.Errorf("%w, host port waiting failed", err)
		}

		if exitCode == 0 {
			break
		} else if exitCode == 126 {
			return errShellNotExecutable
		}
	}
	return nil
}

func buildInternalCheckCommand(internalPort int) string {
	command := `(
					cat /proc/net/tcp* | awk '{print $2}' | grep -i :%04x ||
					nc -vz -w 1 localhost %d ||
					/bin/sh -c '</dev/tcp/localhost/%d'
				)
				`
	return "true && " + fmt.Sprintf(command, internalPort, internalPort, internalPort)
}
