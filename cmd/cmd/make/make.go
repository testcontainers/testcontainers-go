package make

import (
	"github.com/spf13/cobra"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
)

func NewMakeCmd(ctx *pkg_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use: "make",
	}
	cmd.AddCommand(NewAddExampleCmd(ctx))
	return cmd
}
