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
	Long:  "Refresh the module and example files, including the dependabot config, mkdocs config, vscode settings and makefiles for all the modules and examples",
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx, err := context.GetRootContext()
		if err != nil {
			return fmt.Errorf(">> could not get the root dir: %w", err)
		}

		if err := internal.Refresh(ctx); err != nil {
			return fmt.Errorf(">> could not refresh the modules: %w", err)
		}

		fmt.Println("Modules and examples refreshed.")
		fmt.Println("🙏 Commit the modified files and submit a pull request to include them into the project.")
		fmt.Println("Thanks!")
		return nil
	},
}

func init() {
	NewCmd.AddCommand(newExampleCmd)
	NewCmd.AddCommand(newModuleCmd)
}
