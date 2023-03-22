package wait

import (
	"context"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go/exec"
)

func TestWaitForListeningPortSucceeds(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := nat.NewPort("tcp", strconv.Itoa(rawPort))
	if err != nil {
		t.Fatal(err)
	}

	var mappedPortCount, execCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", ErrPortNotFound
			}
			return port, nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
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
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	if err := wg.WaitUntilReady(context.Background(), target); err != nil {
		t.Fatal(err)
	}
}

func TestWaitForExposedPortSucceeds(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := nat.NewPort("tcp", strconv.Itoa(rawPort))
	if err != nil {
		t.Fatal(err)
	}

	var mappedPortCount, execCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		PortsImpl: func(_ context.Context) (nat.PortMap, error) {
			return nat.PortMap{
				"80": []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: port.Port(),
					},
				},
			}, nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", ErrPortNotFound
			}
			return port, nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
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
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	if err := wg.WaitUntilReady(context.Background(), target); err != nil {
		t.Fatal(err)
	}
}

func TestHostPortStrategyFailsWhileGettingPortDueToOOMKilledContainer(t *testing.T) {
	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				OOMKilled: true,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "container crashed with out-of-memory (OOMKilled)"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
}

func TestHostPortStrategyFailsWhileGettingPortDueToExitedContainer(t *testing.T) {
	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "container exited with code 1"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
}

func TestHostPortStrategyFailsWhileGettingPortDueToUnexpectedContainerStatus(t *testing.T) {
	var mappedPortCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			defer func() { mappedPortCount++ }()
			if mappedPortCount == 0 {
				return "", ErrPortNotFound
			}
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status: "dead",
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "unexpected container status \"dead\""
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
}

func TestHostPortStrategyFailsWhileExternalCheckingDueToOOMKilledContainer(t *testing.T) {
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				OOMKilled: true,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "container crashed with out-of-memory (OOMKilled)"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
}

func TestHostPortStrategyFailsWhileExternalCheckingDueToExitedContainer(t *testing.T) {
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "container exited with code 1"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
}

func TestHostPortStrategyFailsWhileExternalCheckingDueToUnexpectedContainerStatus(t *testing.T) {
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return "49152", nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
				Status: "dead",
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "unexpected container status \"dead\""
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
}

func TestHostPortStrategyFailsWhileInternalCheckingDueToOOMKilledContainer(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := nat.NewPort("tcp", strconv.Itoa(rawPort))
	if err != nil {
		t.Fatal(err)
	}

	var stateCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return port, nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			defer func() { stateCount++ }()
			if stateCount == 0 {
				return &types.ContainerState{
					Running: true,
				}, nil
			}
			return &types.ContainerState{
				OOMKilled: true,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "container crashed with out-of-memory (OOMKilled)"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
}

func TestHostPortStrategyFailsWhileInternalCheckingDueToExitedContainer(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := nat.NewPort("tcp", strconv.Itoa(rawPort))
	if err != nil {
		t.Fatal(err)
	}

	var stateCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return port, nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			defer func() { stateCount++ }()
			if stateCount == 0 {
				return &types.ContainerState{
					Running: true,
				}, nil
			}
			return &types.ContainerState{
				Status:   "exited",
				ExitCode: 1,
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "container exited with code 1"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
}

func TestHostPortStrategyFailsWhileInternalCheckingDueToUnexpectedContainerStatus(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := nat.NewPort("tcp", strconv.Itoa(rawPort))
	if err != nil {
		t.Fatal(err)
	}

	var stateCount int
	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return port, nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			defer func() { stateCount++ }()
			if stateCount == 0 {
				return &types.ContainerState{
					Running: true,
				}, nil
			}
			return &types.ContainerState{
				Status: "dead",
			}, nil
		},
	}

	wg := NewHostPortStrategy("80").
		WithStartupTimeout(500 * time.Millisecond).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		if err == nil {
			t.Fatal("no error")
		}

		expected := "unexpected container status \"dead\""
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	}
}
