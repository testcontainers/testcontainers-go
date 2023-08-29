package modules

import (
	"github.com/spf13/cobra"

	"github.com/testcontainers/testcontainers-go/modulegen/internal"
)

var newModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Create a new Module",
	Long:  "Create a new Module",
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.Generate(*exampleVar, true)
	},
}

func init() {
	newModuleCmd.Flags().StringVarP(&exampleVar.Name, "name", "n", "", "Name of the module. Only alphabetical characters are allowed.")
	newModuleCmd.Flags().StringVarP(&exampleVar.NameTitle, "title", "t", "", "(Optional) Title of the module name, used to override the name in the case of mixed casing (Mongodb -> MongoDB). Use camel-case when needed. Only alphabetical characters are allowed.")
	newModuleCmd.Flags().StringVarP(&exampleVar.Image, "image", "i", "", "Fully-qualified name of the Docker image to be used by the module")
	_ = newModuleCmd.MarkFlagRequired("image")
	_ = newModuleCmd.MarkFlagRequired("name")
}
