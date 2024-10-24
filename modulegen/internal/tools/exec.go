package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

func GoModTidy(cmdDir string) error {
	if err := runGoCommand(cmdDir, "mod", "tidy"); err != nil {
		return fmt.Errorf(">> error synchronizing the dependencies: %w", err)
	}
	return nil
}

func GoVet(cmdDir string) error {
	if err := runGoCommand(cmdDir, "vet", "./..."); err != nil {
		return fmt.Errorf(">> error checking generated code: %w", err)
	}
	return nil
}

func GoWorkSync(cmdDir string) error {
	if err := runGoCommand(cmdDir, "work", "sync"); err != nil {
		return fmt.Errorf(">> error syncing work file: %w", err)
	}
	return nil
}

func MakeLint(cmdDir string) error {
	if err := runMakeCommand(cmdDir, "lint"); err != nil {
		return fmt.Errorf(">> error linting: %w", err)
	}
	return nil
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

	var outbuf, errbuf strings.Builder // or bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("[%s] %s %s: %w (%s)", cmdDir, bin, args, err, errbuf.String())
	}
	return nil
}
