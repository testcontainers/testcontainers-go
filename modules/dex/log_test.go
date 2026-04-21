package dex

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlogConsumer_EmitsRecord(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	consumer := newSlogConsumer(logger)

	line := `time="2026-01-01T00:00:00Z" level=warning msg="test message" component=server`
	consumer.accept(line, "STDOUT")

	out := buf.String()
	assert.Contains(t, out, "test message")
	assert.Contains(t, out, "level=WARN")
	assert.Contains(t, out, "component=server")
}

func TestSlogConsumer_StderrMinWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	consumer := newSlogConsumer(logger)

	line := `level=info msg="stderr line"`
	consumer.accept(line, "STDERR")

	out := buf.String()
	require.NotEmpty(t, out)
	assert.Contains(t, out, "level=WARN", "stderr lines promoted to at least WARN")
}

func TestSlogConsumer_EmptyLineIgnored(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	consumer := newSlogConsumer(logger)

	consumer.accept("", "STDOUT")
	consumer.accept("\n", "STDOUT")

	assert.Empty(t, buf.String(), "empty lines must not emit records")
}

func TestParseLogfmt_UnknownKeysBecomeAttrs(t *testing.T) {
	_, msg, attrs := parseLogfmt(`level=info msg=hello foo=bar baz=qux`)
	assert.Equal(t, "hello", msg)
	require.Len(t, attrs, 2)
	assert.Equal(t, "foo", attrs[0].Key)
	assert.Equal(t, "bar", attrs[0].Value.String())
	assert.Equal(t, "baz", attrs[1].Key)
	assert.Equal(t, "qux", attrs[1].Value.String())
}

func TestParseLogfmt_QuotedValue(t *testing.T) {
	_, msg, _ := parseLogfmt(`level=error msg="something went wrong: boom"`)
	assert.Equal(t, "something went wrong: boom", msg)
}

func TestMapLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"warn":    slog.LevelWarn,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
		"fatal":   slog.LevelError,
		"bogus":   slog.LevelInfo, // default
	}
	for in, want := range cases {
		assert.Equal(t, want, mapLevel(in), "input=%q", in)
	}
}
