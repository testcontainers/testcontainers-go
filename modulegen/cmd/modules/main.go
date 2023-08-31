package modules

import (
	"github.com/spf13/cobra"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

var exampleVar = &context.ExampleVar{}

var NewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Example or Module",
	Long:  "Create a new Example or Module",
}

func init() {
	NewCmd.AddCommand(newExampleCmd)
	NewCmd.AddCommand(newModuleCmd)
}
