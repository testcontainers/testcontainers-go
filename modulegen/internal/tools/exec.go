package tools

import (
	"fmt"
	"os/exec"
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

func MakeLint(cmdDir string) error {
	if err := runMakeCommand(cmdDir, "lint"); err != nil {
		return fmt.Errorf(">> error synchronizing the dependencies: %w", err)
	}
	return nil
}

func runGoCommand(cmdDir string, args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Dir = cmdDir
	return cmd.Run()
}

func runMakeCommand(cmdDir string, args ...string) error {
	cmd := exec.Command("make", args...)
	cmd.Dir = cmdDir
	return cmd.Run()
}
