package workflows

import (
	"github.com/spf13/cobra"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
)

func NewWorkflowsCmd(ctx *pkg_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use: "workflows",
	}
	cmd.AddCommand(NewUpdateCmd(ctx))
	return cmd
}
