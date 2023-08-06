package modules

import (
	"fmt"
	"os/exec"
)

func GoModTidy(ctx *Context) error {
	if err := runGoCommand(ctx.RootContext.ExampleDirectory(), "mod", "tidy"); err != nil {
		return fmt.Errorf(">> error synchronizing the dependencies: %w", err)
	}
	return nil
}

func GoVet(ctx *Context) error {
	if err := runGoCommand(ctx.RootContext.ExampleDirectory(), "vet", "./..."); err != nil {
		return fmt.Errorf(">> error checking generated code: %w", err)
	}
	return nil
}

func runGoCommand(cmdDir string, args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Dir = cmdDir
	return cmd.Run()
}
