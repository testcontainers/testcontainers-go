package testcontainers

import (
	"fmt"
	"os/exec"
	"strconv"
)

func fetchComposeMajorVersion(executable string) (int, error) {
	cmd := exec.Command(executable, "version", "--short")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}
	maybeVersion := majorVersionRe.Find(output)
	majorVersion, err := strconv.Atoi(string(maybeVersion))
	if err != nil {
		return majorVersion, fmt.Errorf("%w", err)
	}
	return majorVersion, nil
}
