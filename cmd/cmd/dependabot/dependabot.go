package dependabot

import (
	"github.com/spf13/cobra"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
)

func NewDependabotCmd(ctx *pkg_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use: "dependabot",
	}
	cmd.AddCommand(NewAddExampleCmd(ctx))
	return cmd
}
