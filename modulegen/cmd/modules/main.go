package modules

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/testcontainers/testcontainers-go/modulegen/internal"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

var tcModuleVar = context.TestcontainersModuleVar{}

var NewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Example or Module",
	Long:  "Create a new Example or Module",
}

var RefreshModulesCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the module and example files",
	Long:  "Refresh the module and example files, including the dependabot config, mkdocs config, sonar properties, vscode settings and makefiles for all the modules and examples",
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx, err := context.GetRootContext()
		if err != nil {
			return fmt.Errorf(">> could not get the root dir: %w", err)
		}

		return internal.Refresh(ctx)
	},
}

func init() {
	NewCmd.AddCommand(newExampleCmd)
	NewCmd.AddCommand(newModuleCmd)
}
