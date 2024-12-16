package ollama_test

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types/strslice"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/ollama"
)

func TestRun_local(t *testing.T) {
	// check if the local ollama binary is available
	if _, err := exec.LookPath("ollama"); err != nil {
		t.Skip("local ollama binary not found, skipping")
	}

	ctx := context.Background()

	ollamaContainer, err := ollama.Run(
		ctx,
		"ollama/ollama:0.1.25",
		ollama.WithUseLocal("FOO=BAR"),
	)
	testcontainers.CleanupContainer(t, ollamaContainer)
	require.NoError(t, err)

	t.Run("connection-string", func(t *testing.T) {
		connectionStr, err := ollamaContainer.ConnectionString(ctx)
		require.NoError(t, err)
		require.Equal(t, "http://127.0.0.1:11434", connectionStr)
	})

	t.Run("container-id", func(t *testing.T) {
		id := ollamaContainer.GetContainerID()
		require.Equal(t, "local-ollama-"+testcontainers.SessionID(), id)
	})

	t.Run("container-ips", func(t *testing.T) {
		ip, err := ollamaContainer.ContainerIP(ctx)
		require.NoError(t, err)
		require.Equal(t, "127.0.0.1", ip)

		ips, err := ollamaContainer.ContainerIPs(ctx)
		require.NoError(t, err)
		require.Equal(t, []string{"127.0.0.1"}, ips)
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
		endpoint, err := ollamaContainer.Endpoint(ctx, "88888/tcp")
		require.NoError(t, err)
		require.Equal(t, "127.0.0.1:11434", endpoint)
	})

	t.Run("exec/pull-and-run-model", func(t *testing.T) {
		const model = "llama3.2:1b"

		code, r, err := ollamaContainer.Exec(ctx, []string{"ollama", "pull", model})
		require.NoError(t, err)
		require.Equal(t, 0, code)

		bs, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Empty(t, bs)

		code, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "run", model}, tcexec.Multiplexed())
		require.NoError(t, err)
		require.Equal(t, 0, code)

		logs, err := ollamaContainer.Logs(ctx)
		require.NoError(t, err)
		defer logs.Close()

		bs, err = io.ReadAll(logs)
		require.NoError(t, err)
		require.Contains(t, string(bs), "llama runner started")
	})

	t.Run("exec/unsupported-command", func(t *testing.T) {
		code, r, err := ollamaContainer.Exec(ctx, []string{"cat", "/etc/passwd"})
		require.Equal(t, 1, code)
		require.Error(t, err)
		require.ErrorIs(t, err, errors.ErrUnsupported)

		bs, err := io.ReadAll(r)
		require.NoError(t, err)
		require.Equal(t, "cat: unsupported operation", string(bs))

		code, r, err = ollamaContainer.Exec(ctx, []string{})
		require.Equal(t, 1, code)
		require.Error(t, err)

		bs, err = io.ReadAll(r)
		require.NoError(t, err)
		require.Equal(t, "exec: no command provided", string(bs))
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
		require.Equal(t, "127.0.0.1", host)
	})

	t.Run("inspect", func(t *testing.T) {
		inspect, err := ollamaContainer.Inspect(ctx)
		require.NoError(t, err)

		require.Equal(t, "local-ollama-"+testcontainers.SessionID(), inspect.ContainerJSONBase.ID)
		require.Equal(t, "local-ollama-"+testcontainers.SessionID(), inspect.ContainerJSONBase.Name)
		require.True(t, inspect.ContainerJSONBase.State.Running)

		require.Contains(t, string(inspect.Config.Image), "ollama version is")
		_, exists := inspect.Config.ExposedPorts["11434/tcp"]
		require.True(t, exists)
		require.Equal(t, "localhost", inspect.Config.Hostname)
		require.Equal(t, strslice.StrSlice(strslice.StrSlice{"ollama", "serve"}), inspect.Config.Entrypoint)

		require.Empty(t, inspect.NetworkSettings.Networks)
		require.Equal(t, "bridge", inspect.NetworkSettings.NetworkSettingsBase.Bridge)

		ports := inspect.NetworkSettings.NetworkSettingsBase.Ports
		_, exists = ports["11434/tcp"]
		require.True(t, exists)

		require.Equal(t, "127.0.0.1", inspect.NetworkSettings.Ports["11434/tcp"][0].HostIP)
		require.Equal(t, "11434", inspect.NetworkSettings.Ports["11434/tcp"][0].HostPort)
	})

	t.Run("logfile", func(t *testing.T) {
		openFile, err := os.Open("local-ollama-" + testcontainers.SessionID() + ".log")
		require.NoError(t, err)
		require.NotNil(t, openFile)
		require.NoError(t, openFile.Close())
	})

	t.Run("logs", func(t *testing.T) {
		logs, err := ollamaContainer.Logs(ctx)
		require.NoError(t, err)
		defer logs.Close()

		bs, err := io.ReadAll(logs)
		require.NoError(t, err)

		require.Contains(t, string(bs), "Listening on 127.0.0.1:11434")
	})

	t.Run("mapped-port", func(t *testing.T) {
		port, err := ollamaContainer.MappedPort(ctx, "11434/tcp")
		require.NoError(t, err)
		require.Equal(t, "11434", port.Port())
		require.Equal(t, "tcp", port.Proto())
	})

	t.Run("networks", func(t *testing.T) {
		networks, err := ollamaContainer.Networks(ctx)
		require.NoError(t, err)
		require.Empty(t, networks)
	})

	t.Run("network-aliases", func(t *testing.T) {
		aliases, err := ollamaContainer.NetworkAliases(ctx)
		require.NoError(t, err)
		require.Empty(t, aliases)
	})

	t.Run("session-id", func(t *testing.T) {
		id := ollamaContainer.SessionID()
		require.Equal(t, testcontainers.SessionID(), id)
	})

	t.Run("stop-start", func(t *testing.T) {
		d := time.Second * 5

		err := ollamaContainer.Stop(ctx, &d)
		require.NoError(t, err)

		state, err := ollamaContainer.State(ctx)
		require.NoError(t, err)
		require.Equal(t, "exited", state.Status)

		err = ollamaContainer.Start(ctx)
		require.NoError(t, err)

		state, err = ollamaContainer.State(ctx)
		require.NoError(t, err)
		require.Equal(t, "running", state.Status)

		logs, err := ollamaContainer.Logs(ctx)
		require.NoError(t, err)
		defer logs.Close()

		bs, err := io.ReadAll(logs)
		require.NoError(t, err)

		require.Contains(t, string(bs), "Listening on 127.0.0.1:11434")
	})

	t.Run("start-start", func(t *testing.T) {
		state, err := ollamaContainer.State(ctx)
		require.NoError(t, err)
		require.Equal(t, "running", state.Status)

		err = ollamaContainer.Start(ctx)
		require.NoError(t, err)
	})

	t.Run("terminate", func(t *testing.T) {
		err := ollamaContainer.Terminate(ctx)
		require.NoError(t, err)

		_, err = os.Stat("ollama-" + testcontainers.SessionID() + ".log")
		require.True(t, os.IsNotExist(err))

		state, err := ollamaContainer.State(ctx)
		require.NoError(t, err)
		require.Equal(t, "exited", state.Status)
	})
}

func TestRun_localWithCustomLogFile(t *testing.T) {
	t.Setenv("OLLAMA_LOGFILE", filepath.Join(t.TempDir(), "server.log"))

	ctx := context.Background()

	ollamaContainer, err := ollama.Run(ctx, "ollama/ollama:0.1.25", ollama.WithUseLocal("FOO=BAR"))
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, ollamaContainer)

	logs, err := ollamaContainer.Logs(ctx)
	require.NoError(t, err)
	defer logs.Close()

	bs, err := io.ReadAll(logs)
	require.NoError(t, err)

	require.Contains(t, string(bs), "Listening on 127.0.0.1:11434")
}

func TestRun_localWithCustomHost(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "127.0.0.1:1234")

	ctx := context.Background()

	ollamaContainer, err := ollama.Run(ctx, "ollama/ollama:0.1.25", ollama.WithUseLocal())
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, ollamaContainer)

	t.Run("connection-string", func(t *testing.T) {
		connectionStr, err := ollamaContainer.ConnectionString(ctx)
		require.NoError(t, err)
		require.Equal(t, "http://127.0.0.1:1234", connectionStr)
	})

	t.Run("endpoint", func(t *testing.T) {
		endpoint, err := ollamaContainer.Endpoint(ctx, "1234/tcp")
		require.NoError(t, err)
		require.Equal(t, "127.0.0.1:1234", endpoint)
	})

	t.Run("inspect", func(t *testing.T) {
		inspect, err := ollamaContainer.Inspect(ctx)
		require.NoError(t, err)

		require.Contains(t, string(inspect.Config.Image), "ollama version is")
		_, exists := inspect.Config.ExposedPorts["1234/tcp"]
		require.True(t, exists)
		require.Equal(t, "localhost", inspect.Config.Hostname)
		require.Equal(t, strslice.StrSlice(strslice.StrSlice{"ollama", "serve"}), inspect.Config.Entrypoint)

		require.Empty(t, inspect.NetworkSettings.Networks)
		require.Equal(t, "bridge", inspect.NetworkSettings.NetworkSettingsBase.Bridge)

		ports := inspect.NetworkSettings.NetworkSettingsBase.Ports
		_, exists = ports["1234/tcp"]
		require.True(t, exists)

		require.Equal(t, "127.0.0.1", inspect.NetworkSettings.Ports["1234/tcp"][0].HostIP)
		require.Equal(t, "1234", inspect.NetworkSettings.Ports["1234/tcp"][0].HostPort)
	})

	t.Run("logs", func(t *testing.T) {
		logs, err := ollamaContainer.Logs(ctx)
		require.NoError(t, err)
		defer logs.Close()

		bs, err := io.ReadAll(logs)
		require.NoError(t, err)

		require.Contains(t, string(bs), "Listening on 127.0.0.1:1234")
	})

	t.Run("mapped-port", func(t *testing.T) {
		port, err := ollamaContainer.MappedPort(ctx, "1234/tcp")
		require.NoError(t, err)
		require.Equal(t, "1234", port.Port())
		require.Equal(t, "tcp", port.Proto())
	})
}
