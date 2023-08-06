package mkdocs

import (
	"github.com/spf13/cobra"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
	pkg_mkdocs "github.com/testcontainers/testcontainers-go/cmd/pkg/mkdocs"
)

func NewAddExampleCmd(rootCtx *pkg_cmd.RootContext) *cobra.Command {
	ctx := &pkg_mkdocs.Context{RootContext: rootCtx}
	cmd := &cobra.Command{
		Use: "addExample",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := pkg_mkdocs.GenerateDocFromContext(ctx); err != nil {
				return err
			}
			configFile := rootCtx.MkDocsConfigFile()
			mkdocsConfig, err := pkg_mkdocs.ReadConfig(configFile)
			if err != nil {
				return err
			}
			mkdocsConfig.AddExampleFromContext(ctx)
			return pkg_mkdocs.WriteConfig(configFile, mkdocsConfig)
		},
	}
	cmd.Flags().StringVarP(&rootCtx.Image, "image", "i", "", "Fully-qualified name of the Docker image to be used by the example")
	cmd.MarkFlagRequired("image")
	return cmd
}
