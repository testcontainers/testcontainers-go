package wait

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/exec"
	tclog "github.com/testcontainers/testcontainers-go/log"
)

func TestWaitForListeningPortSucceeds(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := network.ParsePort(fmt.Sprintf("%d/tcp", rawPort))
	require.NoError(t, err)

	var mappedPortCount, execCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return network.Port{}, ErrPortNotFound
			}
			return port, nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Running: true,
			}, nil
		},
		ExecImpl: func(_ context.Context, _ []string, _ ...exec.ProcessOption) (int, io.Reader, error) {
			defer func() { execCount++ }()
			if execCount == 0 {
				return 1, nil, nil
			}
			return 0, nil, nil
		},
	}

	wg := ForListeningPort("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	err = wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

func TestWaitForListeningPortInternallySucceeds(t *testing.T) {
	localPort, err := network.ParsePort("80/tcp")
	require.NoError(t, err)

	mappedPort, err := network.ParsePort("8080/tcp")
	require.NoError(t, err)

	var mappedPortCount, execCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, port string) (network.Port, error) {
			p, err := network.ParsePort(port)
			if err != nil {
				return network.Port{}, err
			}
			if p.Num() != localPort.Num() {
				return network.Port{}, ErrPortNotFound
			}
			defer func() { mappedPortCount++ }()
			if mappedPortCount <= 2 {
				return network.Port{}, ErrPortNotFound
			}
			return mappedPort, nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Running: true,
			}, nil
		},
		ExecImpl: func(_ context.Context, _ []string, _ ...exec.ProcessOption) (int, io.Reader, error) {
			defer func() { execCount++ }()
			if execCount <= 2 {
				return 1, nil, nil
			}
			return 0, nil, nil
		},
	}

	wg := ForListeningPort(localPort.String()).
		SkipExternalCheck().
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	err = wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

func TestWaitForMappedPortSucceeds(t *testing.T) {
	localPort, err := network.ParsePort("80/tcp")
	require.NoError(t, err)

	mappedPort, err := network.ParsePort("8080/tcp")
	require.NoError(t, err)

	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, port string) (network.Port, error) {
			p, err := network.ParsePort(port)
			if err != nil {
				return network.Port{}, err
			}
			if p.Num() != localPort.Num() {
				return network.Port{}, ErrPortNotFound
			}
			defer func() { mappedPortCount++ }()
			if mappedPortCount <= 2 {
				return network.Port{}, ErrPortNotFound
			}
			return mappedPort, nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Running: true,
			}, nil
		},
	}

	wg := ForMappedPort(localPort.String()).
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	err = wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

func TestWaitForExposedPortSkipChecksSucceeds(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := network.ParsePort(fmt.Sprintf("%d/tcp", rawPort))
	require.NoError(t, err)

	var inspectCount, mappedPortCount, execCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			defer func() { inspectCount++ }()
			if inspectCount == 0 {
				// Simulate a container that hasn't bound any ports yet.
				return &container.InspectResponse{
					NetworkSettings: &container.NetworkSettings{
						Ports: network.PortMap{},
					},
				}, nil
			}

			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					Ports: network.PortMap{
						network.MustParsePort("80"): []network.PortBinding{
							{
								HostIP:   netip.MustParseAddr("0.0.0.0"),
								HostPort: port.Port(),
							},
						},
					},
				},
			}, nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return network.Port{}, ErrPortNotFound
			}
			return port, nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Running: true,
			}, nil
		},
		ExecImpl: func(_ context.Context, _ []string, _ ...exec.ProcessOption) (int, io.Reader, error) {
			defer func() { execCount++ }()
			if execCount == 0 {
				return 1, nil, nil
			}
			return 0, nil, nil
		},
	}

	wg := ForExposedPort().
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	err = wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)
}

func TestHostPortStrategyFailsWhileGettingPortDueToOOMKilledContainer(t *testing.T) {
	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return network.Port{}, ErrPortNotFound
			}
			return network.MustParsePort("49152"), nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				OOMKilled: true,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.ErrorContains(t, err, "container crashed with out-of-memory (OOMKilled)")
	}
}

func TestHostPortStrategyFailsWhileGettingPortDueToExitedContainer(t *testing.T) {
	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return network.Port{}, ErrPortNotFound
			}
			return network.MustParsePort("49152"), nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status:   container.StateExited,
				ExitCode: 1,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.ErrorContains(t, err, "container exited with code 1")
	}
}

func TestHostPortStrategyFailsWhileGettingPortDueToUnexpectedContainerStatus(t *testing.T) {
	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return network.Port{}, ErrPortNotFound
			}
			return network.MustParsePort("49152"), nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status: container.StateDead,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.ErrorContains(t, err, "unexpected container status \"dead\"")
	}
}

func TestHostPortStrategyFailsWhileExternalCheckingDueToOOMKilledContainer(t *testing.T) {
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			return network.MustParsePort("49152"), nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				OOMKilled: true,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.ErrorContains(t, err, "container crashed with out-of-memory (OOMKilled)")
	}
}

func TestHostPortStrategyFailsWhileExternalCheckingDueToExitedContainer(t *testing.T) {
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			return network.MustParsePort("49152"), nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status:   container.StateExited,
				ExitCode: 1,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.ErrorContains(t, err, "container exited with code 1")
	}
}

func TestHostPortStrategyFailsWhileExternalCheckingDueToUnexpectedContainerStatus(t *testing.T) {
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			return network.MustParsePort("49152"), nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Status: container.StateDead,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.ErrorContains(t, err, "unexpected container status \"dead\"")
	}
}

func TestHostPortStrategyFailsWhileInternalCheckingDueToOOMKilledContainer(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := network.ParsePort(fmt.Sprintf("%d/tcp", rawPort))
	require.NoError(t, err)

	var stateCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			return port, nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			defer func() { stateCount++ }()
			if stateCount == 0 {
				return &container.State{
					Running: true,
				}, nil
			}
			return &container.State{
				OOMKilled: true,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.ErrorContains(t, err, "container crashed with out-of-memory (OOMKilled)")
	}
}

func TestHostPortStrategyFailsWhileInternalCheckingDueToExitedContainer(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := network.ParsePort(fmt.Sprintf("%d/tcp", rawPort))
	require.NoError(t, err)

	var stateCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			return port, nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			defer func() { stateCount++ }()
			if stateCount == 0 {
				return &container.State{
					Running: true,
				}, nil
			}
			return &container.State{
				Status:   container.StateExited,
				ExitCode: 1,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.ErrorContains(t, err, "container exited with code 1")
	}
}

func TestHostPortStrategyFailsWhileInternalCheckingDueToUnexpectedContainerStatus(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := network.ParsePort(fmt.Sprintf("%d/tcp", rawPort))
	require.NoError(t, err)

	var stateCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			return port, nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			defer func() { stateCount++ }()
			if stateCount == 0 {
				return &container.State{
					Running: true,
				}, nil
			}
			return &container.State{
				Status: container.StateDead,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.ErrorContains(t, err, "unexpected container status \"dead\"")
	}
}

func TestHostPortStrategySucceedsGivenShellIsNotInstalled(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := network.ParsePort(fmt.Sprintf("%d/tcp", rawPort))
	require.NoError(t, err)

	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					Ports: network.PortMap{
						network.MustParsePort("80"): []network.PortBinding{
							{
								HostIP:   netip.MustParseAddr("0.0.0.0"),
								HostPort: port.Port(),
							},
						},
					},
				},
			}, nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			return port, nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Running: true,
			}, nil
		},
		ExecImpl: func(_ context.Context, _ []string, _ ...exec.ProcessOption) (int, io.Reader, error) {
			// This is the error that would be returned if the shell is not installed.
			return exitEaccess, nil, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	oldLogger := tclog.Default()

	var buf bytes.Buffer
	logger := log.New(&buf, "test", log.LstdFlags)

	tclog.SetDefault(logger)
	t.Cleanup(func() {
		tclog.SetDefault(oldLogger)
	})

	err = wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)

	require.Contains(t, buf.String(), "Shell not executable in container, only external port validated")
}

func TestHostPortStrategySucceedsGivenShellIsNotFound(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := network.ParsePort(fmt.Sprintf("%d/tcp", rawPort))
	require.NoError(t, err)

	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		InspectImpl: func(_ context.Context) (*container.InspectResponse, error) {
			return &container.InspectResponse{
				NetworkSettings: &container.NetworkSettings{
					Ports: network.PortMap{
						network.MustParsePort("80"): []network.PortBinding{
							{
								HostIP:   netip.MustParseAddr("0.0.0.0"),
								HostPort: port.Port(),
							},
						},
					},
				},
			}, nil
		},
		MappedPortImpl: func(_ context.Context, _ string) (network.Port, error) {
			return port, nil
		},
		StateImpl: func(_ context.Context) (*container.State, error) {
			return &container.State{
				Running: true,
			}, nil
		},
		ExecImpl: func(_ context.Context, _ []string, _ ...exec.ProcessOption) (int, io.Reader, error) {
			// This is the error that would be returned if the shell is not found.
			return exitCmdNotFound, nil, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	oldLogger := tclog.Default()

	var buf bytes.Buffer
	logger := log.New(&buf, "test", log.LstdFlags)

	tclog.SetDefault(logger)
	t.Cleanup(func() {
		tclog.SetDefault(oldLogger)
	})

	err = wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)

	require.Contains(t, buf.String(), "Shell not found in container")
}
