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
	Short: "Create a new Module",
	Long:  "Create a new Module",
}

var RefreshModulesCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the module files",
	Long:  "Refresh the module files, including the mkdocs config, vscode settings and makefiles for all the modules",
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx, err := context.GetRootContext()
		if err != nil {
			return fmt.Errorf(">> could not get the root dir: %w", err)
		}

		if err := internal.Refresh(ctx); err != nil {
			return fmt.Errorf(">> could not refresh the modules: %w", err)
		}

		fmt.Println("Modules refreshed.")
		fmt.Println("🙏 Commit the modified files and submit a pull request to include them into the project.")
		fmt.Println("Thanks!")
		return nil
	},
}

func init() {
	NewCmd.AddCommand(newModuleCmd)
}
