package modules

import (
	"github.com/spf13/cobra"

	"github.com/testcontainers/testcontainers-go/modulegen/internal"
)

var newExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Create a new Example",
	Long:  "Create a new Example",
	RunE: func(_ *cobra.Command, _ []string) error {
		return internal.Generate(tcModuleVar, false)
	},
}

func init() {
	newExampleCmd.Flags().StringVarP(&tcModuleVar.Name, nameFlag, "n", "", "Name of the example. Only alphabetical characters are allowed.")
	newExampleCmd.Flags().StringVarP(&tcModuleVar.NameTitle, titleFlag, "t", "", "(Optional) Title of the example name, used to override the name in the case of mixed casing (Mongodb -> MongoDB). Use camel-case when needed. Only alphabetical characters are allowed.")
	newExampleCmd.Flags().StringVarP(&tcModuleVar.Image, imageFlag, "i", "", "Fully-qualified name of the Docker image to be used by the example")

	_ = newExampleCmd.MarkFlagRequired(imageFlag)
	_ = newExampleCmd.MarkFlagRequired(nameFlag)
}
