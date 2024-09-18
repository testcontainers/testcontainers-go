package wait

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

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
		WithStartupTimeout(5 * time.Second).
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
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80": []nat.PortBinding{
								{
									HostIP:   "0.0.0.0",
									HostPort: port.Port(),
								},
							},
						},
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
		WithStartupTimeout(5 * time.Second).
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
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)

		expected := "container crashed with out-of-memory (OOMKilled)"
		require.Contains(t, err.Error(), expected)
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
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)

		expected := "container exited with code 1"
		require.Contains(t, err.Error(), expected)
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
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)

		expected := "unexpected container status \"dead\""
		require.Contains(t, err.Error(), expected)
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
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)

		expected := "container crashed with out-of-memory (OOMKilled)"
		require.Contains(t, err.Error(), expected)
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
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)

		expected := "container exited with code 1"
		require.Contains(t, err.Error(), expected)
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
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)

		expected := "unexpected container status \"dead\""
		require.Contains(t, err.Error(), expected)
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
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)

		expected := "container crashed with out-of-memory (OOMKilled)"
		require.Contains(t, err.Error(), expected)
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
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)

		expected := "container exited with code 1"
		require.Contains(t, err.Error(), expected)
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
		WithStartupTimeout(5 * time.Second).
		WithPollInterval(100 * time.Millisecond)

	{
		err := wg.WaitUntilReady(context.Background(), target)
		require.Error(t, err)

		expected := "unexpected container status \"dead\""
		require.Contains(t, err.Error(), expected)
	}
}

func TestHostPortStrategySucceedsGivenShellIsNotInstalled(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	rawPort := listener.Addr().(*net.TCPAddr).Port
	port, err := nat.NewPort("tcp", strconv.Itoa(rawPort))
	require.NoError(t, err)

	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80": []nat.PortBinding{
								{
									HostIP:   "0.0.0.0",
									HostPort: port.Port(),
								},
							},
						},
					},
				},
			}, nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return port, nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
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

	oldWriter := log.Default().Writer()
	var buf bytes.Buffer
	log.Default().SetOutput(&buf)
	t.Cleanup(func() {
		log.Default().SetOutput(oldWriter)
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
	port, err := nat.NewPort("tcp", strconv.Itoa(rawPort))
	require.NoError(t, err)

	target := &MockStrategyTarget{
		HostImpl: func(_ context.Context) (string, error) {
			return "localhost", nil
		},
		InspectImpl: func(_ context.Context) (*types.ContainerJSON, error) {
			return &types.ContainerJSON{
				NetworkSettings: &types.NetworkSettings{
					NetworkSettingsBase: types.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80": []nat.PortBinding{
								{
									HostIP:   "0.0.0.0",
									HostPort: port.Port(),
								},
							},
						},
					},
				},
			}, nil
		},
		MappedPortImpl: func(_ context.Context, _ nat.Port) (nat.Port, error) {
			return port, nil
		},
		StateImpl: func(_ context.Context) (*types.ContainerState, error) {
			return &types.ContainerState{
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

	oldWriter := log.Default().Writer()
	var buf bytes.Buffer
	log.Default().SetOutput(&buf)
	t.Cleanup(func() {
		log.Default().SetOutput(oldWriter)
	})

	err = wg.WaitUntilReady(context.Background(), target)
	require.NoError(t, err)

	require.Contains(t, buf.String(), "Shell not found in container")
}
