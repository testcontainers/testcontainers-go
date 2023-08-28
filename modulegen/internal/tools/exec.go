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

func runGoCommand(cmdDir string, args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Dir = cmdDir
	return cmd.Run()
}
