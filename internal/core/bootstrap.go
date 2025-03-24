package core

import (
	"os"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/testcontainers/testcontainers-go/internal/config"
)

// sessionID returns a unique session ID for the current test session.
var sessionID string

// projectPath returns the current working directory of the parent test process running Testcontainers for Go.
// If it's not possible to get that directory, the library will use the current working directory. If again
// it's not possible to get the current working directory, the library will use a temporary directory.
var projectPath string

// processID returns a unique ID for the current test process. Because each Go package will be run in a separate process,
// we need a way to identify the current test process, in the form of a UUID
var processID string

func init() {
	cfg := config.Read()
	if cfg.SessionID != "" {
		sessionID = cfg.SessionID
	} else {
		sessionID = uuid.New().String()
	}

	processID = uuid.New().String()

	parentPid := os.Getppid()
	fallbackCwd, err := os.Getwd()
	if err != nil {
		// very unlikely to fail, but if it does, we will use a temp dir
		fallbackCwd = os.TempDir()
	}

	processes, err := process.Processes()
	if err != nil {
		projectPath = fallbackCwd
		return
	}

	for _, p := range processes {
		if int(p.Pid) != parentPid {
			continue
		}

		cwd, err := p.Cwd()
		if err != nil {
			cwd = fallbackCwd
		}
		projectPath = cwd

		break
	}
}

func ProcessID() string {
	return processID
}

func ProjectPath() string {
	return projectPath
}

func SessionID() string {
	return sessionID
}
