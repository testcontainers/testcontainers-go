package cmd

import (
	"github.com/spf13/cobra"

	"github.com/testcontainers/testcontainers-go/modulegen/cmd/modules"
)

var NewRootCmd = &cobra.Command{
	Use:   "modulegen",
	Short: "Management tool for testcontainers-go",
	Long:  "Management tool for testcontainers-go",
}

func init() {
	NewRootCmd.AddCommand(modules.NewCmd)
	NewRootCmd.AddCommand(modules.RefreshModulesCmd)
}
