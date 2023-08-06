package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/testcontainers/testcontainers-go/cmd/cmd/dependabot"
	"github.com/testcontainers/testcontainers-go/cmd/cmd/make"
	"github.com/testcontainers/testcontainers-go/cmd/cmd/mkdocs"
	"github.com/testcontainers/testcontainers-go/cmd/cmd/modules"
	"github.com/testcontainers/testcontainers-go/cmd/cmd/workflows"
	pkg_cmd "github.com/testcontainers/testcontainers-go/cmd/pkg/cmd"
)

func NewRootCmd() *cobra.Command {
	ctx := &pkg_cmd.RootContext{}
	cmd := &cobra.Command{
		Use: "testcontainers-go",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := ctx.Validate(); err != nil {
				return err
			}
			currentDir, err := filepath.Abs(filepath.Dir("."))
			if err != nil {
				return fmt.Errorf(">> could not get the root dir: %w", err)
			}
			ctx.RootDir = filepath.Dir(currentDir)
			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&ctx.Name, "name", "n", "", "Name of the example. Only alphanumerical characters are allowed.")
	cmd.PersistentFlags().StringVarP(&ctx.Title, "title", "t", "", "(Optional) Title of the example name, used to override the name in the case of mixed casing (Mongodb -> MongoDB). Use camel-case when needed. Only alphanumerical characters are allowed.")
	cmd.PersistentFlags().BoolVar(&ctx.IsModule, "as-module", false, "If set, the example will be generated as a Go module, under the modules directory. Otherwise, it will be generated as a subdirectory of the examples directory.")

	cmd.MarkFlagRequired("name")

	cmd.AddCommand(dependabot.NewDependabotCmd(ctx))
	cmd.AddCommand(mkdocs.NewMkDocsCmd(ctx))
	cmd.AddCommand(workflows.NewWorkflowsCmd(ctx))
	cmd.AddCommand(modules.NewModulesCmd(ctx))
	cmd.AddCommand(make.NewMakeCmd(ctx))
	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
