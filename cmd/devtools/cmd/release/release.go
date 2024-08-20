package release

import (
	"github.com/spf13/cobra"

	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
	"github.com/testcontainers/testcontainers-go/devtools/internal/git"
	"github.com/testcontainers/testcontainers-go/devtools/internal/release"
)

var (
	bumpType   string
	dryRun     bool
	preRelease bool
)

var ReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Performs a release of the Testcontainers Go library",
	Long: `This script is used to prepare a release for a new version of the Testcontainers for Go library.
If the dry-run flag is set, it will be run in dry-run mode, which will print the commands that would be executed, without actually
executing them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := parseBumpType(bumpType); err != nil {
			return err
		}

		ctx, err := context.GetRootContext()
		if err != nil {
			return err
		}

		branch := "main"

		// when using the CLI, we are going to use the main branch for the release
		releaser := release.NewReleaseManager(branch, bumpType, dryRun)

		gitClient := git.New(ctx, branch, dryRun)

		err = releaser.PreRun(ctx, gitClient)
		if err != nil {
			return err
		}

		if preRelease {
			return nil
		}

		return releaser.Run(ctx, gitClient)
	},
}

func init() {
	ReleaseCmd.Flags().BoolVarP(&dryRun, dryRunFlag, "d", false, "If true, the release will be a dry-run and no changes will be made to the repository")
	ReleaseCmd.Flags().BoolVarP(&preRelease, preReleaseFlag, "p", false, "If true, only the pre-release steps will be executed.")
	ReleaseCmd.Flags().StringVarP(&bumpType, bumpTypeFlag, "B", "minor", "The type of bump to perform. "+bumpTypeInfoMsg)
}
