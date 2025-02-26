package tools

import (
	"fmt"
	"os/exec"
)

// GoModTidy synchronizes the dependencies for a module or example,
// running the `go mod tidy` command in the given directory.
func GoModTidy(cmdDir string) error {
	if err := runCommand(cmdDir, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf(">> error synchronizing the dependencies: %w", err)
	}
	return nil
}

// GoVet checks the generated code for errors,
// running the `go vet ./...` command in the given directory.
func GoVet(cmdDir string) error {
	if err := runCommand(cmdDir, "go", "vet", "./..."); err != nil {
		return fmt.Errorf(">> error checking generated code: %w", err)
	}
	return nil
}

// MakeLint runs the `make lint` command in the given directory.
func MakeLint(cmdDir string) error {
	if err := runCommand(cmdDir, "make", "lint"); err != nil {
		return fmt.Errorf(">> error running make lint: %w", err)
	}
	return nil
}

func runCommand(cmdDir string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = cmdDir
	return cmd.Run()
}
