package modules

import (
	"github.com/spf13/cobra"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
	pkg_mkdocs "github.com/testcontainers/testcontainers-go/cmd/pkg/mkdocs"
	pkg_modules "github.com/testcontainers/testcontainers-go/cmd/pkg/modules"
)

func NewInitCmd(rootCtx *pkg_cmd.RootContext) *cobra.Command {
	ctx := &pkg_modules.Context{RootContext: rootCtx}
	cmd := &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := rootCtx.MkDocsConfigFile()
			mkdocsConfig, err := pkg_mkdocs.ReadConfig(configFile)
			if err != nil {
				return err
			}
			ctx.TCVersion = mkdocsConfig.Extra.LatestVersion
			if err := pkg_modules.GenerateFilesFromContext(ctx); err != nil {
				return err
			}
			if err := pkg_modules.GenerateGomod(ctx); err != nil {
				return err
			}
			if err := pkg_modules.GoModTidy(ctx); err != nil {
				return err
			}
			if err := pkg_modules.GoVet(ctx); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&rootCtx.Image, "image", "i", "", "Fully-qualified name of the Docker image to be used by the example")
	cmd.MarkFlagRequired("image")
	return cmd
}
