package ollama_test

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/ollama"
)

const (
	testImage   = "ollama/ollama:latest"
	testNatPort = "11434/tcp"
	testHost    = "127.0.0.1"
	testBinary  = "ollama"
)

var (
	// reLogDetails matches the log details of the local ollama binary and should match localLogRegex.
	reLogDetails = regexp.MustCompile(`Listening on (.*:\d+) \(version\s(.*)\)`)
	zeroTime     = time.Time{}.Format(time.RFC3339Nano)
)

func TestRun_local(t *testing.T) {
	// check if the local ollama binary is available
	if _, err := exec.LookPath(testBinary); err != nil {
		t.Skip("local ollama binary not found, skipping")
	}

	ctx := context.Background()
	ollamaContainer, err := ollama.Run(
		ctx,
		testImage,
		ollama.WithUseLocal("FOO=BAR"),
	)
	testcontainers.CleanupContainer(t, ollamaContainer)
	require.NoError(t, err)

	t.Run("state", func(t *testing.T) {
		state, err := ollamaContainer.State(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, state.StartedAt)
		require.NotEqual(t, zeroTime, state.StartedAt)
		require.NotZero(t, state.Pid)
		require.Equal(t, &container.State{
			Status:     "running",
			Running:    true,
			Pid:        state.Pid,
			StartedAt:  state.StartedAt,
			FinishedAt: state.FinishedAt,
		}, state)
	})

	t.Run("connection-string", func(t *testing.T) {
		connectionStr, err := ollamaContainer.ConnectionString(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, connectionStr)
	})

	t.Run("container-id", func(t *testing.T) {
		id := ollamaContainer.GetContainerID()
		require.Equal(t, "local-ollama-"+testcontainers.SessionID(), id)
	})

	t.Run("container-ips", func(t *testing.T) {
		ip, err := ollamaContainer.ContainerIP(ctx)
		require.NoError(t, err)
		require.Equal(t, testHost, ip)

		ips, err := ollamaContainer.ContainerIPs(ctx)
		require.NoError(t, err)
		require.Equal(t, []string{testHost}, ips)
	})

	t.Run("copy", func(t *testing.T) {
		err := ollamaContainer.CopyToContainer(ctx, []byte("test"), "/tmp", 0o755)
		require.Error(t, err)

		err = ollamaContainer.CopyDirToContainer(ctx, ".", "/tmp", 0o755)
		require.Error(t, err)

		err = ollamaContainer.CopyFileToContainer(ctx, ".", "/tmp", 0o755)
		require.Error(t, err)

		reader, err := ollamaContainer.CopyFileFromContainer(ctx, "/tmp")
		require.Error(t, err)
		require.Nil(t, reader)
	})

	t.Run("log-production-error-channel", func(t *testing.T) {
		ch := ollamaContainer.GetLogProductionErrorChannel()
		require.Nil(t, ch)
	})

	t.Run("endpoint", func(t *testing.T) {
		endpoint, err := ollamaContainer.Endpoint(ctx, "")
		require.NoError(t, err)
		require.Contains(t, endpoint, testHost+":")

		endpoint, err = ollamaContainer.Endpoint(ctx, "http")
		require.NoError(t, err)
		require.Contains(t, endpoint, "http://"+testHost+":")
	})

	t.Run("is-running", func(t *testing.T) {
		require.True(t, ollamaContainer.IsRunning())

		err = ollamaContainer.Stop(ctx, nil)
		require.NoError(t, err)
		require.False(t, ollamaContainer.IsRunning())

		// return it to the running state
		err = ollamaContainer.Start(ctx)
		require.NoError(t, err)
		require.True(t, ollamaContainer.IsRunning())
	})

	t.Run("host", func(t *testing.T) {
		host, err := ollamaContainer.Host(ctx)
		require.NoError(t, err)
		require.Equal(t, testHost, host)
	})

	t.Run("inspect", func(t *testing.T) {
		inspect, err := ollamaContainer.Inspect(ctx)
		require.NoError(t, err)

		require.Equal(t, "local-ollama-"+testcontainers.SessionID(), inspect.ID)
		require.Equal(t, "local-ollama-"+testcontainers.SessionID(), inspect.Name)
		require.True(t, inspect.State.Running)

		require.NotEmpty(t, inspect.Config.Image)
		_, exists := inspect.Config.ExposedPorts[testNatPort]
		require.True(t, exists)
		require.Equal(t, testHost, inspect.Config.Hostname)
		require.Equal(t, strslice.StrSlice(strslice.StrSlice{testBinary, "serve"}), inspect.Config.Entrypoint)

		require.Empty(t, inspect.NetworkSettings.Networks)
		require.Equal(t, "bridge", inspect.NetworkSettings.Bridge)

		ports := inspect.NetworkSettings.Ports
		port, exists := ports[testNatPort]
		require.True(t, exists)
		require.Len(t, port, 1)
		require.Equal(t, testHost, port[0].HostIP)
		require.NotEmpty(t, port[0].HostPort)
	})

	t.Run("logfile", func(t *testing.T) {
		file, err := os.Open("local-ollama-" + testcontainers.SessionID() + ".log")
		require.NoError(t, err)
		require.NoError(t, file.Close())
	})

	t.Run("logs", func(t *testing.T) {
		logs, err := ollamaContainer.Logs(ctx)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, logs.Close())
		})

		bs, err := io.ReadAll(logs)
		require.NoError(t, err)
		require.Regexp(t, reLogDetails, string(bs))
	})

	t.Run("mapped-port", func(t *testing.T) {
		port, err := ollamaContainer.MappedPort(ctx, testNatPort)
		require.NoError(t, err)
		require.NotEmpty(t, port.Port())
		require.Equal(t, "tcp", port.Proto())
	})

	t.Run("networks", func(t *testing.T) {
		networks, err := ollamaContainer.Networks(ctx)
		require.NoError(t, err)
		require.Nil(t, networks)
	})

	t.Run("network-aliases", func(t *testing.T) {
		aliases, err := ollamaContainer.NetworkAliases(ctx)
		require.NoError(t, err)
		require.Nil(t, aliases)
	})

	t.Run("port-endpoint", func(t *testing.T) {
		endpoint, err := ollamaContainer.PortEndpoint(ctx, testNatPort, "")
		require.NoError(t, err)
		require.Regexp(t, `^127.0.0.1:\d+$`, endpoint)

		endpoint, err = ollamaContainer.PortEndpoint(ctx, testNatPort, "http")
		require.NoError(t, err)
		require.Regexp(t, `^http://127.0.0.1:\d+$`, endpoint)
	})

	t.Run("session-id", func(t *testing.T) {
		require.Equal(t, testcontainers.SessionID(), ollamaContainer.SessionID())
	})

	t.Run("stop-start", func(t *testing.T) {
		d := time.Second * 5
		err := ollamaContainer.Stop(ctx, &d)
		require.NoError(t, err)

		state, err := ollamaContainer.State(ctx)
		require.NoError(t, err)
		require.Equal(t, "exited", state.Status)
		require.NotEmpty(t, state.StartedAt)
		require.NotEqual(t, zeroTime, state.StartedAt)
		require.NotEmpty(t, state.FinishedAt)
		require.NotEqual(t, zeroTime, state.FinishedAt)
		require.Zero(t, state.ExitCode)

		err = ollamaContainer.Start(ctx)
		require.NoError(t, err)

		state, err = ollamaContainer.State(ctx)
		require.NoError(t, err)
		require.Equal(t, "running", state.Status)

		logs, err := ollamaContainer.Logs(ctx)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, logs.Close())
		})

		bs, err := io.ReadAll(logs)
		require.NoError(t, err)
		require.Regexp(t, reLogDetails, string(bs))
	})

	t.Run("start-start", func(t *testing.T) {
		state, err := ollamaContainer.State(ctx)
		require.NoError(t, err)
		require.Equal(t, "running", state.Status)

		err = ollamaContainer.Start(ctx)
		require.Error(t, err)
	})

	t.Run("terminate", func(t *testing.T) {
		err := ollamaContainer.Terminate(ctx)
		require.NoError(t, err)

		_, err = os.Stat("ollama-" + testcontainers.SessionID() + ".log")
		require.ErrorIs(t, err, fs.ErrNotExist)

		state, err := ollamaContainer.State(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, state.StartedAt)
		require.NotEqual(t, zeroTime, state.StartedAt)
		require.NotEmpty(t, state.FinishedAt)
		require.NotEqual(t, zeroTime, state.FinishedAt)
		require.Equal(t, &container.State{
			// zero values are not needed to be set
			Status:     "exited",
			StartedAt:  state.StartedAt,
			FinishedAt: state.FinishedAt,
		}, state)
	})

	t.Run("deprecated", func(t *testing.T) {
		t.Run("ports", func(t *testing.T) {
			inspect, err := ollamaContainer.Inspect(ctx)
			require.NoError(t, err)

			ports, err := ollamaContainer.Ports(ctx)
			require.NoError(t, err)
			require.Equal(t, inspect.NetworkSettings.Ports, ports)
		})

		t.Run("follow-output", func(t *testing.T) {
			require.Panics(t, func() {
				ollamaContainer.FollowOutput(&testcontainers.StdoutLogConsumer{})
			})
		})

		t.Run("start-log-producer", func(t *testing.T) {
			err := ollamaContainer.StartLogProducer(ctx)
			require.ErrorIs(t, err, errors.ErrUnsupported)
		})

		t.Run("stop-log-producer", func(t *testing.T) {
			err := ollamaContainer.StopLogProducer()
			require.ErrorIs(t, err, errors.ErrUnsupported)
		})

		t.Run("name", func(t *testing.T) {
			name, err := ollamaContainer.Name(ctx)
			require.NoError(t, err)
			require.Equal(t, "local-ollama-"+testcontainers.SessionID(), name)
		})
	})
}

func TestRun_localWithCustomLogFile(t *testing.T) {
	ctx := context.Background()
	logFile := filepath.Join(t.TempDir(), "server.log")

	t.Run("parent-env", func(t *testing.T) {
		t.Setenv("OLLAMA_LOGFILE", logFile)

		ollamaContainer, err := ollama.Run(ctx, testImage, ollama.WithUseLocal())
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.NoError(t, err)

		logs, err := ollamaContainer.Logs(ctx)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, logs.Close())
		})

		bs, err := io.ReadAll(logs)
		require.NoError(t, err)
		require.Regexp(t, reLogDetails, string(bs))

		file, ok := logs.(*os.File)
		require.True(t, ok)
		require.Equal(t, logFile, file.Name())
	})

	t.Run("local-env", func(t *testing.T) {
		ollamaContainer, err := ollama.Run(ctx, testImage, ollama.WithUseLocal("OLLAMA_LOGFILE="+logFile))
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.NoError(t, err)

		logs, err := ollamaContainer.Logs(ctx)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, logs.Close())
		})

		bs, err := io.ReadAll(logs)
		require.NoError(t, err)
		require.Regexp(t, reLogDetails, string(bs))

		file, ok := logs.(*os.File)
		require.True(t, ok)
		require.Equal(t, logFile, file.Name())
	})
}

func TestRun_localWithCustomHost(t *testing.T) {
	ctx := context.Background()

	t.Run("parent-env", func(t *testing.T) {
		t.Setenv("OLLAMA_HOST", "127.0.0.1:1234")

		ollamaContainer, err := ollama.Run(ctx, testImage, ollama.WithUseLocal())
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.NoError(t, err)

		testRunLocalWithCustomHost(ctx, t, ollamaContainer)
	})

	t.Run("local-env", func(t *testing.T) {
		ollamaContainer, err := ollama.Run(ctx, testImage, ollama.WithUseLocal("OLLAMA_HOST=127.0.0.1:1234"))
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.NoError(t, err)

		testRunLocalWithCustomHost(ctx, t, ollamaContainer)
	})
}

func testRunLocalWithCustomHost(ctx context.Context, t *testing.T, ollamaContainer *ollama.OllamaContainer) {
	t.Helper()

	t.Run("connection-string", func(t *testing.T) {
		connectionStr, err := ollamaContainer.ConnectionString(ctx)
		require.NoError(t, err)
		require.Equal(t, "http://127.0.0.1:1234", connectionStr)
	})

	t.Run("endpoint", func(t *testing.T) {
		endpoint, err := ollamaContainer.Endpoint(ctx, "http")
		require.NoError(t, err)
		require.Equal(t, "http://127.0.0.1:1234", endpoint)
	})

	t.Run("inspect", func(t *testing.T) {
		inspect, err := ollamaContainer.Inspect(ctx)
		require.NoError(t, err)
		require.Regexp(t, `^local-ollama:\d+\.\d+\.\d+$`, inspect.Config.Image)

		_, exists := inspect.Config.ExposedPorts[testNatPort]
		require.True(t, exists)
		require.Equal(t, testHost, inspect.Config.Hostname)
		require.Equal(t, strslice.StrSlice(strslice.StrSlice{testBinary, "serve"}), inspect.Config.Entrypoint)

		require.Empty(t, inspect.NetworkSettings.Networks)
		require.Equal(t, "bridge", inspect.NetworkSettings.Bridge)

		ports := inspect.NetworkSettings.Ports
		port, exists := ports[testNatPort]
		require.True(t, exists)
		require.Len(t, port, 1)
		require.Equal(t, testHost, port[0].HostIP)
		require.Equal(t, "1234", port[0].HostPort)
	})

	t.Run("logs", func(t *testing.T) {
		logs, err := ollamaContainer.Logs(ctx)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, logs.Close())
		})

		bs, err := io.ReadAll(logs)
		require.NoError(t, err)

		require.Contains(t, string(bs), "Listening on 127.0.0.1:1234")
	})

	t.Run("mapped-port", func(t *testing.T) {
		port, err := ollamaContainer.MappedPort(ctx, testNatPort)
		require.NoError(t, err)
		require.Equal(t, "1234", port.Port())
		require.Equal(t, "tcp", port.Proto())
	})
}

func TestRun_localExec(t *testing.T) {
	// check if the local ollama binary is available
	if _, err := exec.LookPath(testBinary); err != nil {
		t.Skip("local ollama binary not found, skipping")
	}

	ctx := context.Background()

	ollamaContainer, err := ollama.Run(ctx, testImage, ollama.WithUseLocal())
	testcontainers.CleanupContainer(t, ollamaContainer)
	require.NoError(t, err)

	t.Run("no-command", func(t *testing.T) {
		code, r, err := ollamaContainer.Exec(ctx, nil)
		require.Error(t, err)
		require.Equal(t, 1, code)
		require.Nil(t, r)
	})

	t.Run("unsupported-command", func(t *testing.T) {
		code, r, err := ollamaContainer.Exec(ctx, []string{"cat", "/etc/hosts"})
		require.ErrorIs(t, err, errors.ErrUnsupported)
		require.Equal(t, 1, code)
		require.Nil(t, r)
	})

	t.Run("unsupported-option-user", func(t *testing.T) {
		code, r, err := ollamaContainer.Exec(ctx, []string{testBinary, "-v"}, tcexec.WithUser("root"))
		require.ErrorIs(t, err, errors.ErrUnsupported)
		require.Equal(t, 1, code)
		require.Nil(t, r)
	})

	t.Run("unsupported-option-privileged", func(t *testing.T) {
		code, r, err := ollamaContainer.Exec(ctx, []string{testBinary, "-v"}, tcexec.ProcessOptionFunc(func(opts *tcexec.ProcessOptions) {
			opts.ExecConfig.Privileged = true
		}))
		require.ErrorIs(t, err, errors.ErrUnsupported)
		require.Equal(t, 1, code)
		require.Nil(t, r)
	})

	t.Run("unsupported-option-tty", func(t *testing.T) {
		code, r, err := ollamaContainer.Exec(ctx, []string{testBinary, "-v"}, tcexec.ProcessOptionFunc(func(opts *tcexec.ProcessOptions) {
			opts.ExecConfig.Tty = true
		}))
		require.ErrorIs(t, err, errors.ErrUnsupported)
		require.Equal(t, 1, code)
		require.Nil(t, r)
	})

	t.Run("unsupported-option-detach", func(t *testing.T) {
		code, r, err := ollamaContainer.Exec(ctx, []string{testBinary, "-v"}, tcexec.ProcessOptionFunc(func(opts *tcexec.ProcessOptions) {
			opts.ExecConfig.Detach = true
		}))
		require.ErrorIs(t, err, errors.ErrUnsupported)
		require.Equal(t, 1, code)
		require.Nil(t, r)
	})

	t.Run("unsupported-option-detach-keys", func(t *testing.T) {
		code, r, err := ollamaContainer.Exec(ctx, []string{testBinary, "-v"}, tcexec.ProcessOptionFunc(func(opts *tcexec.ProcessOptions) {
			opts.ExecConfig.DetachKeys = "ctrl-p,ctrl-q"
		}))
		require.ErrorIs(t, err, errors.ErrUnsupported)
		require.Equal(t, 1, code)
		require.Nil(t, r)
	})

	t.Run("pull-and-run-model", func(t *testing.T) {
		const model = "llama3.2:1b"

		code, r, err := ollamaContainer.Exec(ctx, []string{testBinary, "pull", model})
		require.NoError(t, err)
		require.Zero(t, code)

		bs, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Contains(t, string(bs), "success")

		code, r, err = ollamaContainer.Exec(ctx, []string{testBinary, "run", model}, tcexec.Multiplexed())
		require.NoError(t, err)
		require.Zero(t, code)

		bs, err = io.ReadAll(r)
		require.NoError(t, err)
		require.Empty(t, bs)

		logs, err := ollamaContainer.Logs(ctx)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, logs.Close())
		})

		bs, err = io.ReadAll(logs)
		require.NoError(t, err)
		require.Contains(t, string(bs), "llama runner started")
	})
}

func TestRun_localValidateRequest(t *testing.T) {
	// check if the local ollama binary is available
	if _, err := exec.LookPath(testBinary); err != nil {
		t.Skip("local ollama binary not found, skipping")
	}

	ctx := context.Background()
	t.Run("waiting-for-nil", func(t *testing.T) {
		ollamaContainer, err := ollama.Run(
			ctx,
			testImage,
			ollama.WithUseLocal("FOO=BAR"),
			testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
				req.WaitingFor = nil
				return nil
			}),
		)
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.EqualError(t, err, "validate request: ContainerRequest.WaitingFor must be set")
	})

	t.Run("started-false", func(t *testing.T) {
		ollamaContainer, err := ollama.Run(
			ctx,
			testImage,
			ollama.WithUseLocal("FOO=BAR"),
			testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
				req.Started = false
				return nil
			}),
		)
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.EqualError(t, err, "validate request: started must be true")
	})

	t.Run("exposed-ports-empty", func(t *testing.T) {
		ollamaContainer, err := ollama.Run(
			ctx,
			testImage,
			ollama.WithUseLocal("FOO=BAR"),
			testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
				req.ExposedPorts = req.ExposedPorts[:0]
				return nil
			}),
		)
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.EqualError(t, err, "validate request: ContainerRequest.ExposedPorts must be 11434/tcp got: []")
	})

	t.Run("dockerfile-set", func(t *testing.T) {
		ollamaContainer, err := ollama.Run(
			ctx,
			testImage,
			ollama.WithUseLocal("FOO=BAR"),
			testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
				req.Dockerfile = "FROM scratch"
				return nil
			}),
		)
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.EqualError(t, err, "validate request: unsupported field: ContainerRequest.FromDockerfile.Dockerfile = \"FROM scratch\"")
	})

	t.Run("image-only", func(t *testing.T) {
		ollamaContainer, err := ollama.Run(
			ctx,
			testBinary,
			ollama.WithUseLocal(),
		)
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.NoError(t, err)
	})

	t.Run("image-path", func(t *testing.T) {
		ollamaContainer, err := ollama.Run(
			ctx,
			"prefix-path/"+testBinary,
			ollama.WithUseLocal(),
		)
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.NoError(t, err)
	})

	t.Run("image-bad-version", func(t *testing.T) {
		ollamaContainer, err := ollama.Run(
			ctx,
			testBinary+":bad-version",
			ollama.WithUseLocal(),
		)
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.EqualError(t, err, `validate request: ContainerRequest.Image version must be blank or "latest", got: "bad-version"`)
	})

	t.Run("image-not-found", func(t *testing.T) {
		ollamaContainer, err := ollama.Run(
			ctx,
			"ollama/ollama-not-found",
			ollama.WithUseLocal(),
		)
		testcontainers.CleanupContainer(t, ollamaContainer)
		require.EqualError(t, err, `validate request: invalid image "ollama/ollama-not-found": exec: "ollama-not-found": executable file not found in $PATH`)
	})
}
