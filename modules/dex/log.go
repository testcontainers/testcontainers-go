package dex

import (
	"context"
	"log/slog"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

// slogConsumer adapts a *slog.Logger to testcontainers.LogConsumer. Dex
// emits logfmt-style lines (key=value); we parse level + msg and preserve
// remaining fields as slog attrs.
//
// Stderr lines are promoted to at least slog.LevelWarn because Dex writes
// runtime errors there.
type slogConsumer struct {
	logger *slog.Logger
}

// Compile check: *slogConsumer implements testcontainers.LogConsumer.
var _ testcontainers.LogConsumer = (*slogConsumer)(nil)

func newSlogConsumer(l *slog.Logger) *slogConsumer {
	return &slogConsumer{logger: l}
}

// Accept implements testcontainers.LogConsumer.
func (s *slogConsumer) Accept(l testcontainers.Log) {
	s.accept(string(l.Content), l.LogType)
}

// accept is the testable inner method. It takes the raw content string and
// the testcontainers log type (STDOUT/STDERR) so unit tests don't need to
// construct a real tc-go Log value.
func (s *slogConsumer) accept(content, logType string) {
	line := strings.TrimRight(content, "\n")
	if line == "" {
		return
	}
	level, msg, attrs := parseLogfmt(line)
	if logType == "STDERR" && level < slog.LevelWarn {
		level = slog.LevelWarn
	}
	s.logger.LogAttrs(context.Background(), level, msg, attrs...)
}

// logfmtUnescaper unescapes the two backslash sequences logfmt allows
// inside quoted values: \" → " and \\ → \.
var logfmtUnescaper = strings.NewReplacer(`\\`, `\`, `\"`, `"`)

// parseLogfmt is a minimal logfmt parser — enough for Dex's default
// format (level=... msg=...). Unknown keys become slog attrs. Quoted
// values are unquoted and have their \" / \\ escapes expanded.
func parseLogfmt(line string) (slog.Level, string, []slog.Attr) {
	level := slog.LevelInfo
	msg := ""
	var attrs []slog.Attr

	pairs := tokenizeLogfmt(line)
	for _, p := range pairs {
		switch p.key {
		case "level":
			level = mapLevel(p.val)
		case "msg":
			msg = p.val
		case "time":
			// Dex timestamps are redundant — slog adds its own.
		default:
			attrs = append(attrs, slog.String(p.key, p.val))
		}
	}
	if msg == "" {
		msg = line
	}
	return level, msg, attrs
}

type kv struct{ key, val string }

func tokenizeLogfmt(line string) []kv {
	var out []kv
	i := 0
	for i < len(line) {
		for i < len(line) && line[i] == ' ' {
			i++
		}
		if i >= len(line) {
			break
		}
		kStart := i
		for i < len(line) && line[i] != '=' && line[i] != ' ' {
			i++
		}
		if i >= len(line) || line[i] != '=' {
			out = append(out, kv{line[kStart:i], ""})
			continue
		}
		k := line[kStart:i]
		i++ // skip '='
		if i < len(line) && line[i] == '"' {
			i++
			vStart := i
			for i < len(line) && line[i] != '"' {
				if line[i] == '\\' && i+1 < len(line) {
					i++
				}
				i++
			}
			out = append(out, kv{k, logfmtUnescaper.Replace(line[vStart:i])})
			if i < len(line) {
				i++
			}
		} else {
			vStart := i
			for i < len(line) && line[i] != ' ' {
				i++
			}
			out = append(out, kv{k, line[vStart:i]})
		}
	}
	return out
}

func mapLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "fatal":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
