package cubrid

import (
	"errors"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	startupTimeout = 2 * time.Minute
	failureLines   = 5
)

func startupStrategy() *wait.LogStrategy {
	return wait.ForLog(`(?s).+`).
		Submatch(func(_ string, matches [][][]byte) error {
			if len(matches) == 0 {
				return errors.New("no output yet")
			}
			return checkStartup(string(matches[0][0]))
		}).
		WithStartupTimeout(startupTimeout)
}

func checkStartup(output string) error {
	lines := make([]string, 0, strings.Count(output, "\n")+1)
	for _, line := range strings.Split(output, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}

	var broker, server bool
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "Couldn't create database."),
			strings.HasPrefix(line, "ERROR:"),
			strings.HasPrefix(line, "++ ") && strings.HasSuffix(line, ": fail"):
			end := min(i+failureLines, len(lines))
			return wait.NewPermanentError(errors.New("cubrid failed to start:\n" + strings.Join(lines[i:end], "\n")))
		case line == "++ cubrid broker start: success":
			broker = true
		case line == "++ cubrid server start: success":
			server = true
		}
	}

	if !broker || !server {
		return errors.New("cubrid broker and server have not started yet")
	}
	return nil
}
