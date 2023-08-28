package main

import (
	"github.com/testcontainers/testcontainers-go/modulegen/internal/workflow"
)

// update github ci workflow
func generateWorkFlow(ctx *Context) error {
	rootCtx, err := getRootContext()
	if err != nil {
		return err
	}
	examples, err := rootCtx.GetExamples()
	if err != nil {
		return err
	}
	modules, err := rootCtx.GetModules()
	if err != nil {
		return err
	}
	return workflow.Generate(ctx.GithubWorkflowsDir(), examples, modules)
}
