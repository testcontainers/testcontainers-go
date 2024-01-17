package tools

import (
	"os/exec"
)

func GoModTidy(cmdDir string) error {
	return runGoCommand(cmdDir, "mod", "tidy")
}

func GoVet(cmdDir string) error {
	return runGoCommand(cmdDir, "vet", "./...")
}

func MakeLint(cmdDir string) error {
	return runMakeCommand(cmdDir, "lint")
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
