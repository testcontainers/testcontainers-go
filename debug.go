package testcontainers

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

// testLogConsumer is a testcontainers LogConsumer and io.Writer
// that writes to the test log.
type testLogConsumer struct {
	t      *testing.T
	prefix string
}

// NewTestLogConsumer creates a new Logger for t.
func NewTestLogConsumer(t *testing.T, prefix string) *testLogConsumer {
	t.Helper()

	return &testLogConsumer{t: t, prefix: prefix}
}

// Accept implements LogConsumer.
func (l *testLogConsumer) Accept(log Log) {
	l.t.Log(time.Now().Format(time.RFC3339Nano), l.prefix, log.LogType, strings.TrimSpace(string(log.Content)))
}

// Write implements io.Write.
func (l *testLogConsumer) Write(p []byte) (int, error) {
	l.t.Log(time.Now().Format(time.RFC3339Nano), l.prefix, strings.TrimSpace(string(p)))
	return len(p), nil
}

var debugEnabledTime time.Time

// DebugTest applies debugging to t which:
//   - Logs reaper container output to the test.
//   - Enables docker debugging and outputs the docker daemon logs before and after the test to the test.
func DebugTest(t *testing.T) {
	t.Helper()
	config.Reset()
	// t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	t.Setenv("TESTCONTAINERS_RYUK_VERBOSE", "true")

	oldDebugPrintln := DebugPrintln
	DebugPrintln = func(a ...any) {
		t.Log(append([]any{time.Now().Format(time.RFC3339Nano)}, a...)...)
	}
	t.Cleanup(func() {
		DebugPrintln = oldDebugPrintln
		config.Reset()
	})

	// Stream reaper logs
	container, err := spawner.lookupContainer(context.Background(), core.SessionID())
	if err != nil {
		t.Logf("look up container: %s", err)
		return
	}

	log := NewTestLogConsumer(t, "reaper:")
	timeout := time.Hour
	container.logProductionTimeout = &timeout
	container.logProductionCtx, container.logProductionCancel = context.WithCancelCause(context.Background())
	t.Cleanup(func() {
		container.logProductionCancel(errLogProductionStop)
	})

	go func() {
		if err := container.logProducer(log, log); err != nil {
			t.Logf("error running logProducer: %s", err)
		}
	}()

	DebugDocker(t, true)

	t.Cleanup(func() {
		DebugDocker(t, false)
	})
}

// Docker enables or disables Docker debug logging and outputs
// the logs to t before and after the test so there is a full trace
// of the docker actions performed.
func DebugDocker(t *testing.T, enable bool) {
	t.Helper()

	t.Log("Docker debug logging:", enable)

	const file = "/etc/docker/daemon.json"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Logf("error reading daemon.json: %s", err)
		return
	}

	t.Logf("daemon.json: %s", string(data))

	cfg := make(map[string]any)
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		t.Logf("error unmarshalling daemon.json: %s", err)
		return
	}

	cfg["debug"] = enable

	data, err = json.Marshal(cfg)
	if err != nil {
		t.Logf("error marshalling daemon.json: %s", err)
		return
	}

	t.Logf("daemon.json: %s", string(data))

	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Logf("error writing daemon.json: %s", err)
		return
	}

	defer os.Remove(f.Name())

	if _, err := f.Write(data); err != nil {
		t.Logf("error writing daemon.json: %s", err)
		return
	}

	cmd := exec.CommandContext(context.Background(), "sudo", "cp", f.Name(), file)
	log, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("error restarting docker: %s, %s", err, log)
		return
	}

	if enable {
		debugEnabledTime = time.Now()
	}
	cmd = exec.CommandContext(context.Background(), "sudo", "systemctl", "reload-or-restart", "docker")
	log, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("error reloading docker: %s, %s", err, log)
		cmd = exec.CommandContext(context.Background(), "journalctl", "-xeu", "docker.service")
		log, err = cmd.CombinedOutput()
		if err != nil {
			t.Logf("error running journalctl: %s, %s", err, log)
			return
		}
		t.Logf("docker journalctl: %s", log)
		return
	}

	t.Logf("docker reloaded: %s", log)
}

// DebugContainerFilter returns a filter for the given container.
func DebugContainerFilter(container Container) []string {
	return []string{
		"type=container",
		"container=" + container.GetContainerID(),
	}
}

// DebugInfo details of the following to the test log:
//   - Docker events for the container.
//   - Docker version and info.
//   - Docker logs if docker debugging is enabled.
func DebugInfo(t *testing.T, filter ...string) {
	t.Helper()

	// Docker events.
	time.Sleep(time.Second) // Events are not immediately available.
	args := []string{"events"}
	for _, v := range filter {
		args = append(args, "--filter", v)
	}
	args = append(args, "--since", "10m", "--until", "0s")
	cmd := exec.CommandContext(context.Background(), "docker", args...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("error running docker events: %s", err)
	} else {
		t.Logf("docker events: %s", stdoutStderr)
	}

	// Docker version.
	cmd = exec.CommandContext(context.Background(), "docker", "version")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("error running docker version: %s", err)
	} else {
		t.Logf("docker version: %s", stdoutStderr)
	}

	// Docker info.
	cmd = exec.CommandContext(context.Background(), "docker", "info")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("error running docker info: %s", err)
	} else {
		t.Logf("docker info: %s", stdoutStderr)
	}

	// Docker logs.
	if debugEnabledTime.IsZero() {
		t.Log("debugEnabledTime is zero, skipping journalctl")
	} else {
		cmd = exec.CommandContext(context.Background(),
			"journalctl", "-xu", "docker.service",
			"--since", debugEnabledTime.Format("2006-01-02 15:04:05"),
		)
		stdoutStderr, err = cmd.CombinedOutput()
		if err != nil {
			t.Logf("error running journalctl: %s, %s", err, stdoutStderr)
		} else {
			t.Logf("docker journalctl: %s", stdoutStderr)
		}
		debugEnabledTime = time.Time{}
	}
}
