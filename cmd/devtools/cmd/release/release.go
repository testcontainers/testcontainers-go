package release

import (
	"github.com/spf13/cobra"

	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
	"github.com/testcontainers/testcontainers-go/devtools/internal/release"
)

var (
	branch string
	dryRun bool
)

var ReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Performs a release of the Testcontainers Go library",
	Long: `This script is used to prepare a release for a new version of the Testcontainers for Go library.
If the dry-run flag is set, it will be run in dry-run mode, which will print the commands that would be executed, without actually
executing them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := context.GetRootContext()
		if err != nil {
			return err
		}

		releaser := release.NewReleaseManager(branch, dryRun)

		return releaser.PreRun(ctx)
	},
}

func init() {
	ReleaseCmd.Flags().BoolVarP(&dryRun, dryRunFlag, "d", false, "If true, the release will be a dry-run and no changes will be made to the repository")
	ReleaseCmd.Flags().StringVarP(&branch, branchFlag, "b", "main", "The branch to perform the release on. Default is 'main'")
}
