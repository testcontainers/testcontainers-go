package main

import (
	"github.com/testcontainers/testcontainers-go/modulegen/internal/dependabot"
)

// update examples in dependabot
func generateDependabotUpdates(ctx *Context, example Example) error {
	directory := "/" + example.ParentDir() + "/" + example.Lower()
	return dependabot.UpdateConfig(ctx.DependabotConfigFile(), directory, "gomod")
}
