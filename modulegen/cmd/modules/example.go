package modules

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/testcontainers/testcontainers-go/modulegen/internal"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

var newExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Create a new Example",
	Long:  "Create a new Example",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := context.GetRootContext()
		if err != nil {
			return fmt.Errorf(">> could not get the root dir: %w", err)
		}

		tcModule := context.TestcontainersModule{
			Image:     tcModuleVar.Image,
			IsModule:  false,
			Name:      tcModuleVar.Name,
			TitleName: tcModuleVar.NameTitle,
		}

		return internal.Generate(ctx, tcModule)
	},
}

func init() {
	newExampleCmd.Flags().StringVarP(&tcModuleVar.Name, nameFlag, "n", "", "Name of the example. Only alphabetical characters are allowed.")
	newExampleCmd.Flags().StringVarP(&tcModuleVar.NameTitle, titleFlag, "t", "", "(Optional) Title of the example name, used to override the name in the case of mixed casing (Mongodb -> MongoDB). Use camel-case when needed. Only alphabetical characters are allowed.")
	newExampleCmd.Flags().StringVarP(&tcModuleVar.Image, imageFlag, "i", "", "Fully-qualified name of the Docker image to be used by the example")

	_ = newExampleCmd.MarkFlagRequired(imageFlag)
	_ = newExampleCmd.MarkFlagRequired(nameFlag)
}
