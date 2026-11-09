package testcontainers

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/log"
)

// fakeLogFrame is a log entry served by fakeLogsDaemon.
type fakeLogFrame struct {
	stream    stdcopy.StdType
	timestamp string
	content   string
}

// fakeLogsResponse is the response fakeLogsDaemon serves to one logs request.
type fakeLogsResponse struct {
	frames []fakeLogFrame

	// fail aborts the connection after the frames instead of ending the stream.
	fail bool
}

// fakeLogsDaemon is a Docker API server that serves canned responses to logs requests.
type fakeLogsDaemon struct {
	mtx       sync.Mutex
	responses []fakeLogsResponse
	sinces    []string
}

// ServeHTTP serves the next canned response and records the since query parameter.
func (d *fakeLogsDaemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/logs") {
		http.NotFound(w, r)
		return
	}

	d.mtx.Lock()
	if len(d.responses) == 0 {
		d.mtx.Unlock()
		http.Error(w, "unexpected logs request", http.StatusInternalServerError)
		return
	}
	resp := d.responses[0]
	d.responses = d.responses[1:]
	d.sinces = append(d.sinces, r.URL.Query().Get("since"))
	d.mtx.Unlock()

	w.Header().Set("Content-Type", "application/vnd.docker.multiplexed-stream")
	w.WriteHeader(http.StatusOK)
	for _, frame := range resp.frames {
		entry := frame.content
		if frame.timestamp != "" {
			entry = frame.timestamp + " " + entry
		}
		header := make([]byte, 8)
		header[0] = byte(frame.stream)
		binary.BigEndian.PutUint32(header[4:], uint32(len(entry)))
		_, _ = w.Write(append(header, entry...))
	}
	w.(http.Flusher).Flush()

	if resp.fail {
		// Close the connection without ending the chunked body.
		panic(http.ErrAbortHandler)
	}
}

// Sinces returns the since query parameter of each logs request received.
func (d *fakeLogsDaemon) Sinces() []string {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	return d.sinces
}

// recordingLogConsumer records every log it receives.
type recordingLogConsumer struct {
	mtx  sync.Mutex
	logs []string
}

// Accept implements LogConsumer.
func (c *recordingLogConsumer) Accept(l Log) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.logs = append(c.logs, l.LogType+": "+string(l.Content))
}

// runFakeLogProducer runs the log producer against a fakeLogsDaemon serving responses,
// returning the logs consumed and the since parameter of each logs request.
func runFakeLogProducer(t *testing.T, responses ...fakeLogsResponse) (logs, sinces []string) {
	t.Helper()

	daemon := &fakeLogsDaemon{responses: responses}
	srv := httptest.NewServer(daemon)
	t.Cleanup(srv.Close)

	cli, err := client.New(client.WithHost("tcp://"+srv.Listener.Addr().String()), client.WithAPIVersion("1.44"))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	c := &DockerContainer{
		ID:               "fake",
		provider:         &DockerProvider{client: cli},
		logger:           log.TestLogger(t),
		logProductionCtx: ctx,
	}

	consumer := &recordingLogConsumer{}
	consumers := []LogConsumer{consumer}
	c.logProducer(newLogConsumerWriter(StdoutLog, consumers), newLogConsumerWriter(StderrLog, consumers))
	require.NoError(t, ctx.Err(), "log producer didn't finish")

	return consumer.logs, daemon.Sinces()
}

// sinceQuery returns the since query parameter the Docker client sends for timestamp.
func sinceQuery(t *testing.T, timestamp string) string {
	t.Helper()

	ts, err := time.Parse(time.RFC3339Nano, timestamp)
	require.NoError(t, err)

	return fmt.Sprintf("%d.%09d", ts.Unix(), ts.Nanosecond())
}

func TestLogProducerResumesAfterError(t *testing.T) {
	t.Run("docker", func(t *testing.T) {
		// Docker sends nanosecond timestamps and since is inclusive.
		const (
			ts1 = "2026-11-09T12:00:00.000000001Z"
			ts2 = "2026-11-09T12:00:00.000000002Z"
			ts3 = "2026-11-09T12:00:00.000000003Z"
		)

		logs, sinces := runFakeLogProducer(t,
			fakeLogsResponse{
				frames: []fakeLogFrame{
					{stdcopy.Stdout, ts1, "one\n"},
					{stdcopy.Stderr, ts2, "two\n"},
				},
				fail: true,
			},
			fakeLogsResponse{
				frames: []fakeLogFrame{
					{stdcopy.Stderr, ts2, "two\n"},
					{stdcopy.Stdout, ts2, "three\n"},
					{stdcopy.Stdout, ts3, "four\n"},
				},
			},
		)

		require.Equal(t, []string{"", sinceQuery(t, ts2)}, sinces)
		require.Equal(t, []string{
			"STDOUT: one\n",
			"STDERR: two\n",
			"STDOUT: three\n",
			"STDOUT: four\n",
		}, logs)
	})

	t.Run("podman", func(t *testing.T) {
		// Podman's Docker compatible API sends timestamps to the second, so resuming
		// replays every entry in the cursor's second, including identical ones.
		const (
			ts1 = "2026-11-09T12:00:00+00:00"
			ts2 = "2026-11-09T12:00:01+00:00"
		)

		logs, sinces := runFakeLogProducer(t,
			fakeLogsResponse{
				frames: []fakeLogFrame{
					{stdcopy.Stdout, ts1, "same\n"},
					{stdcopy.Stderr, ts1, "same\n"},
					{stdcopy.Stdout, ts1, "same\n"},
				},
				fail: true,
			},
			fakeLogsResponse{
				frames: []fakeLogFrame{
					{stdcopy.Stdout, ts1, "same\n"},
					{stdcopy.Stderr, ts1, "same\n"},
					{stdcopy.Stdout, ts1, "same\n"},
					{stdcopy.Stdout, ts1, "same\n"},
					{stdcopy.Stderr, ts2, "last\n"},
				},
			},
		)

		require.Equal(t, []string{"", sinceQuery(t, ts1)}, sinces)
		require.Equal(t, []string{
			"STDOUT: same\n",
			"STDERR: same\n",
			"STDOUT: same\n",
			"STDOUT: same\n",
			"STDERR: last\n",
		}, logs)
	})

	t.Run("no-timestamps", func(t *testing.T) {
		// Entries without a timestamp are passed through unchanged.
		logs, sinces := runFakeLogProducer(t,
			fakeLogsResponse{
				frames: []fakeLogFrame{
					{stdcopy.Stdout, "", "plain\n"},
					{stdcopy.Stderr, "", "not-a-timestamp entry\n"},
				},
			},
		)

		require.Equal(t, []string{""}, sinces)
		require.Equal(t, []string{
			"STDOUT: plain\n",
			"STDERR: not-a-timestamp entry\n",
		}, logs)
	})
}

func TestSplitLogTimestamp(t *testing.T) {
	tests := []struct {
		entry     string
		timestamp string
		content   string
		ok        bool
	}{
		{entry: "2026-11-09T12:00:00.123456789Z msg\n", timestamp: "2026-11-09T12:00:00.123456789Z", content: "msg\n", ok: true},
		{entry: "2026-11-09T12:00:00-05:00 msg\n", timestamp: "2026-11-09T12:00:00-05:00", content: "msg\n", ok: true},
		{entry: "2026-11-09T12:00:00Z \n", timestamp: "2026-11-09T12:00:00Z", content: "\n", ok: true},
		{entry: "msg\n", content: "msg\n"},
		{entry: "2026-11-09 12:00:00 msg\n", content: "2026-11-09 12:00:00 msg\n"},
		{entry: "a short entry\n", content: "a short entry\n"},
	}

	for _, tc := range tests {
		t.Run(strings.TrimSpace(tc.entry), func(t *testing.T) {
			timestamp, content, ok := splitLogTimestamp([]byte(tc.entry))
			require.Equal(t, tc.ok, ok)
			require.Equal(t, tc.timestamp, string(timestamp))
			require.Equal(t, tc.content, string(content))
		})
	}
}

func BenchmarkTimestampLogWriter(b *testing.B) {
	entries := make([][]byte, 1024)
	start := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	for i := range entries {
		ts := start.Add(time.Duration(i) * time.Microsecond).Format("2006-01-02T15:04:05.000000000Z07:00")
		entries[i] = []byte(ts + " 2026/11/09 12:00:00 [info] request handled in 3ms status=200 path=/api/v1/items\n")
	}

	w := &timestampLogWriter{writer: io.Discard, cursor: &logCursor{}}
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		_, _ = w.Write(entries[i%len(entries)])
	}
}
