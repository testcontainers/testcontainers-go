//go:build !freebsd && !openbsd && !windows

package bootstrap

import (
	"github.com/shirou/gopsutil/v4/process"
)

// findParent returns the parent process, or nil if the process list does not include it.
// It returns an error if the process list cannot be read. It reads the list of process IDs
// and then only the parent process, not every process on the machine.
func findParent(parentPid int) (*process.Process, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}

	for _, pid := range pids {
		if int(pid) != parentPid {
			continue
		}

		p, err := process.NewProcess(pid)
		if err != nil {
			return nil, nil
		}

		return p, nil
	}

	return nil, nil
}
