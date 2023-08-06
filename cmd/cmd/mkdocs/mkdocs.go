package mkdocs

import (
	"github.com/spf13/cobra"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
)

func NewMkDocsCmd(ctx *pkg_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use: "mkdocs",
	}
	cmd.AddCommand(NewAddExampleCmd(ctx))
	return cmd
}
