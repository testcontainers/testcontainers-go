package dependabot

import (
	"github.com/spf13/cobra"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
	pkg_dependabot "github.com/testcontainers/testcontainers-go/cmd/pkg/dependabot"
)

func NewAddExampleCmd(rootCtx *pkg_cmd.RootContext) *cobra.Command {
	ctx := &pkg_dependabot.Context{RootContext: rootCtx}
	cmd := &cobra.Command{
		Use: "addExample",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := ctx.ConfigFile()
			dependabotConfig, err := pkg_dependabot.ReadConfig(configFile)
			if err != nil {
				return err
			}
			dependabotConfig.AddExampleFromContext(ctx)
			return pkg_dependabot.WriteConfig(configFile, dependabotConfig)
		},
	}
	return cmd
}
