package modules

import (
	"github.com/spf13/cobra"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
)

func NewModulesCmd(ctx *pkg_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use: "modules",
	}
	cmd.AddCommand(NewInitCmd(ctx))
	return cmd
}
