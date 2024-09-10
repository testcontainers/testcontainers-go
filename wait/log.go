package wait

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Implement interface
var (
	_ Strategy        = (*LogStrategy)(nil)
	_ StrategyTimeout = (*LogStrategy)(nil)
)

// LogStrategy will wait until a given log entry shows up in the docker logs
type LogStrategy struct {
	// all Strategies should have a startupTimeout to avoid waiting infinitely
	timeout *time.Duration

	// additional properties
	Log          string
	IsRegexp     bool
	Occurrence   int
	PollInterval time.Duration
}

// NewLogStrategy constructs with polling interval of 100 milliseconds and startup timeout of 60 seconds by default
func NewLogStrategy(log string) *LogStrategy {
	return &LogStrategy{
		Log:          log,
		IsRegexp:     false,
		Occurrence:   1,
		PollInterval: defaultPollInterval(),
	}
}

// fluent builders for each property
// since go has neither covariance nor generics, the return type must be the type of the concrete implementation
// this is true for all properties, even the "shared" ones like startupTimeout

// AsRegexp can be used to change the default behavior of the log strategy to use regexp instead of plain text
func (ws *LogStrategy) AsRegexp() *LogStrategy {
	ws.IsRegexp = true
	return ws
}

// WithStartupTimeout can be used to change the default startup timeout
func (ws *LogStrategy) WithStartupTimeout(timeout time.Duration) *LogStrategy {
	ws.timeout = &timeout
	return ws
}

// WithPollInterval can be used to override the default polling interval of 100 milliseconds
func (ws *LogStrategy) WithPollInterval(pollInterval time.Duration) *LogStrategy {
	ws.PollInterval = pollInterval
	return ws
}

func (ws *LogStrategy) WithOccurrence(o int) *LogStrategy {
	// the number of occurrence needs to be positive
	if o <= 0 {
		o = 1
	}
	ws.Occurrence = o
	return ws
}

// ForLog is the default construction for the fluid interface.
//
// For Example:
//
//	wait.
//		ForLog("some text").
//		WithPollInterval(1 * time.Second)
func ForLog(log string) *LogStrategy {
	return NewLogStrategy(log)
}

func (ws *LogStrategy) Timeout() *time.Duration {
	return ws.timeout
}

// WaitUntilReady implements Strategy.WaitUntilReady
func (ws *LogStrategy) WaitUntilReady(ctx context.Context, target StrategyTarget) error {
	timeout := defaultStartupTimeout()
	if ws.timeout != nil {
		timeout = *ws.timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var buf bytes.Buffer
	check := ws.checkFunc()
	for {
		found, err := readLogs(ctx, &buf, check, target)
		if err != nil {
			return err
		}

		if found {
			if err := checkTarget(ctx, target); err != nil {
				return err
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for log: %w", ctx.Err())
		case <-time.After(ws.PollInterval):
			if err := checkTarget(ctx, target); err != nil {
				return err
			}
		}
	}
}

// readLogs reads the logs from the target line by line and checks if the log entry is present.
// It returns true if the log entry is found, false otherwise.
// Logs are read until the log entry is found or the context is canceled.
func readLogs(ctx context.Context, buf *bytes.Buffer, check func(buf *bytes.Buffer) bool, target StrategyTarget) (bool, error) {
	reader, err := target.Logs(ctx)
	if err != nil {
		return false, fmt.Errorf("read logs: %w", err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		buf.Write(scanner.Bytes())
		buf.WriteByte('\n')
		if check(buf) {
			return true, nil
		}

		if ctx.Err() != nil {
			return false, fmt.Errorf("scan logs: %w", ctx.Err())
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("scan logs: %w", err)
	}

	return false, nil
}

// checkFunc returns a function that checks if the buffer contains the log entry.
func (ws *LogStrategy) checkFunc() func(buf *bytes.Buffer) bool {
	if ws.IsRegexp {
		re := regexp.MustCompile(ws.Log)
		return func(buf *bytes.Buffer) bool {
			occurrences := re.FindAll(buf.Bytes(), ws.Occurrence)

			return len(occurrences) == ws.Occurrence
		}
	}

	return func(buf *bytes.Buffer) bool {
		return strings.Count(buf.String(), ws.Log) >= ws.Occurrence
	}
}
