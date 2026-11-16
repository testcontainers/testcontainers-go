//go:build freebsd || openbsd || windows

package bootstrap

import (
	"github.com/shirou/gopsutil/v4/process"
)

// findParent returns the parent process, or nil if the process list does not include it.
// It returns an error if the process list cannot be read. On FreeBSD and OpenBSD,
// process.Pids does not return that error, and on Windows process.Processes can fail
// where process.Pids does not, so the list is read with process.Processes.
func findParent(parentPid int) (*process.Process, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	for _, p := range processes {
		if int(p.Pid) == parentPid {
			return p, nil
		}
	}

	return nil, nil
}
