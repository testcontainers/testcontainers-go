package modules

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/testcontainers/testcontainers-go/modulegen/internal"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

var newModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Create a new Module",
	Long:  "Create a new Module",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := context.GetRootContext()
		if err != nil {
			return fmt.Errorf(">> could not get the root dir: %w", err)
		}

		tcModule := context.TestcontainersModule{
			Image:     tcModuleVar.Image,
			IsModule:  true,
			Name:      tcModuleVar.Name,
			TitleName: tcModuleVar.NameTitle,
		}

		return internal.Generate(ctx, tcModule)
	},
}

func init() {
	newModuleCmd.Flags().StringVarP(&tcModuleVar.Name, nameFlag, "n", "", "Name of the module. Only alphabetical characters are allowed.")
	newModuleCmd.Flags().StringVarP(&tcModuleVar.NameTitle, titleFlag, "t", "", "(Optional) Title of the module name, used to override the name in the case of mixed casing (Mongodb -> MongoDB). Use camel-case when needed. Only alphabetical characters are allowed.")
	newModuleCmd.Flags().StringVarP(&tcModuleVar.Image, imageFlag, "i", "", "Fully-qualified name of the Docker image to be used by the module")

	_ = newModuleCmd.MarkFlagRequired(imageFlag)
	_ = newModuleCmd.MarkFlagRequired(nameFlag)
}
