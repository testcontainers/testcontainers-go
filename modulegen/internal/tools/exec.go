package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

func GoModTidy(cmdDir string) error {
	return runGoCommand(cmdDir, "mod", "tidy")
}

func GoVet(cmdDir string) error {
	return runGoCommand(cmdDir, "vet", "./...")
}

func GoWorkSync(cmdDir string) error {
	return runGoCommand(cmdDir, "work", "sync")
}

func MakeLint(cmdDir string) error {
	return runMakeCommand(cmdDir, "lint")
}

func runGoCommand(cmdDir string, args ...string) error {
	return runCommand(cmdDir, "go", args...)
}

func runMakeCommand(cmdDir string, args ...string) error {
	return runCommand(cmdDir, "make", args...)
}

func runCommand(cmdDir string, bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = cmdDir

	var outbuf, errbuf strings.Builder
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("run [%s] %s %s: %w (%s)", cmdDir, bin, args, err, errbuf.String())
	}
	return nil
}
