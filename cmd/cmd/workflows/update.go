package workflows

import (
	"github.com/spf13/cobra"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
	pkg_workflows "github.com/testcontainers/testcontainers-go/cmd/pkg/workflows"
)

func NewUpdateCmd(rootCtx *pkg_cmd.RootContext) *cobra.Command {
	ctx := &pkg_workflows.Context{RootContext: rootCtx}
	cmd := &cobra.Command{
		Use: "update",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg_workflows.GenerateFromContext(ctx)
		},
	}
	return cmd
}
