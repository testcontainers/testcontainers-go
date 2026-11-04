package testcbt

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

// Image is the official full Cloud SDK 592.0.0 image (cbt 1.25.7), pinned
// to its Linux amd64/arm64 OCI index. Only the test helper uses this image.
const Image = "gcr.io/google.com/cloudsdktool/google-cloud-cli:592.0.0@sha256:f81c8c688f09b97b5c78c80e732517f88584eafa5eff50d4ae6b2a5025a37079"

// Limit the command's output at the producer, before Docker Exec buffers it.
// All data is passed as positional arguments, never interpolated into this
// fixed script. Coreutils timeout also bounds the container process itself.
const boundedCommand = `output=$(mktemp) || exit 1
trap 'rm -f "$output"' EXIT
ulimit -f 128 || exit 1
timeout --signal=KILL 25s cbt "$@" >"$output" 2>&1
status=$?
cat "$output" || exit 1
exit "$status"
`

type executor interface {
	Exec(context.Context, []string, ...tcexec.ProcessOption) (int, io.Reader, error)
}

// Run executes one CLI operation, with anonymous emulator transport and
// explicit flags that prevent cbt's gcloud config-helper/credential fallback.
func Run(ctx context.Context, helper executor, endpoint, project, instance string, command ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	args := make([]string, 0, 13+len(command))
	args = append(args,
		"/bin/sh", "-c", boundedCommand, "cbt-test",
		"-project="+project, "-instance="+instance, "-timeout=20s",
		"-admin-endpoint="+endpoint, "-data-endpoint="+endpoint,
		"-creds=/tmp/unused-emulator-credentials",
		"-cert-file=", "-access-token=", "-auth-token=",
	)
	args = append(args, command...)
	code, reader, err := helper.Exec(ctx, args,
		tcexec.WithEnv([]string{"BIGTABLE_EMULATOR_HOST=" + endpoint, "TZ=UTC"}),
		tcexec.Multiplexed(),
	)
	if err != nil {
		return "", fmt.Errorf("cbt %q exec: %w", command, err)
	}
	if reader == nil {
		return "", fmt.Errorf("cbt %q exit %d: missing output reader", command, code)
	}
	const maxOutput = 128 << 10
	data, readErr := io.ReadAll(io.LimitReader(reader, maxOutput+1))
	if readErr != nil {
		return "", fmt.Errorf("cbt %q exit %d: read output: %w", command, code, readErr)
	}
	if len(data) > maxOutput {
		return "", fmt.Errorf("cbt %q exit %d: output exceeds %d bytes", command, code, maxOutput)
	}
	output := string(data)
	if code != 0 {
		return output, fmt.Errorf("cbt %q exit %d: %s", command, code, output)
	}
	return output, nil
}

// Endpoint uses the existing SSH host-access bridge to reach the runner relay.
func Endpoint(port int) string {
	return fmt.Sprintf("%s:%d", testcontainers.HostInternal, port)
}

// LookupCell parses the single row/cell emitted by this test's lookup command.
// Parsing is deliberately limited to the pinned CLI's default quoted format.
func LookupCell(output string) (row, column, value string, err error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 || lines[0] != strings.Repeat("-", 40) {
		return "", "", "", fmt.Errorf("cbt lookup: expected one row and one cell: %q", output)
	}
	row, err = unquote(lines[1])
	if err != nil {
		return "", "", "", err
	}
	index := strings.LastIndex(lines[2], " @ ")
	if index < 0 {
		return "", "", "", fmt.Errorf("cbt lookup: missing column timestamp: %q", lines[2])
	}
	column, err = unquote(lines[2][:index])
	if err != nil {
		return "", "", "", err
	}
	if _, err = time.Parse("2006/01/02-15:04:05.000000", lines[2][index+3:]); err != nil {
		return "", "", "", fmt.Errorf("cbt lookup timestamp: %w", err)
	}
	value, err = unquote(lines[3])
	return row, column, value, err
}

func unquote(line string) (string, error) {
	value, err := strconv.Unquote(strings.TrimSpace(line))
	if err != nil {
		return "", fmt.Errorf("cbt lookup quoted record %q: %w", line, err)
	}
	return value, nil
}
