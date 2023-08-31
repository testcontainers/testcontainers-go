package modules

import (
	"github.com/spf13/cobra"

	"github.com/testcontainers/testcontainers-go/modulegen/internal"
)

var newExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Create a new Example",
	Long:  "Create a new Example",
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.Generate(*exampleVar, false)
	},
}

func init() {
	newExampleCmd.Flags().StringVarP(&exampleVar.Name, nameFlag, "n", "", "Name of the example. Only alphabetical characters are allowed.")
	newExampleCmd.Flags().StringVarP(&exampleVar.NameTitle, titleFlag, "t", "", "(Optional) Title of the example name, used to override the name in the case of mixed casing (Mongodb -> MongoDB). Use camel-case when needed. Only alphabetical characters are allowed.")
	newExampleCmd.Flags().StringVarP(&exampleVar.Image, imageFlag, "i", "", "Fully-qualified name of the Docker image to be used by the example")

	_ = newExampleCmd.MarkFlagRequired(imageFlag)
	_ = newExampleCmd.MarkFlagRequired(nameFlag)
}
