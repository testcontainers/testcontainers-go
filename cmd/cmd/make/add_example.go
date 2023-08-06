package make

import (
	"github.com/spf13/cobra"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
	pkg_make "github.com/testcontainers/testcontainers-go/cmd/pkg/make"
)

func NewAddExampleCmd(rootCtx *pkg_cmd.RootContext) *cobra.Command {
	ctx := &pkg_make.Context{RootContext: rootCtx}
	cmd := &cobra.Command{
		Use: "addExample",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg_make.GenerateFromContext(ctx)
		},
	}
	return cmd
}
