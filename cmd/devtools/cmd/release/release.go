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
	Long:  "Performs a release of the Testcontainers Go library",
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
