package release

import (
	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
	"github.com/testcontainers/testcontainers-go/devtools/internal/git"
)

const (
	golangProxy     string = "https://proxy.golang.org"
	nonReleasedText string = `Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>`
)

var directories = []string{"examples", "modules"}

type Releaser interface {
	PreRun(ctx context.Context, g *git.GitClient) error
	Run(ctx context.Context, g *git.GitClient) error
}

type dryRunReleaseManager struct {
	releaseManager
}

func NewReleaseManager(branch string, bumpType string, dryRun bool) Releaser {
	r := releaseManager{
		branch:   branch,
		bumpType: bumpType,
	}

	if dryRun {
		return &dryRunReleaseManager{
			releaseManager: r,
		}
	}

	return &r
}

func (p *dryRunReleaseManager) PreRun(ctx context.Context, gitClient *git.GitClient) error {
	return preRun(ctx, gitClient, p.branch, true)
}

func (p *dryRunReleaseManager) Run(ctx context.Context, gitClient *git.GitClient) error {
	// dry run and skip remote operations
	return run(ctx, gitClient, p.bumpType, true, true, golangProxy)
}

type releaseManager struct {
	branch   string
	bumpType string
}

func (p *releaseManager) PreRun(ctx context.Context, gitClient *git.GitClient) error {
	return preRun(ctx, gitClient, p.branch, false)
}

func (p *releaseManager) Run(ctx context.Context, gitClient *git.GitClient) error {
	// no dry run and execute remote operations
	return run(ctx, gitClient, p.bumpType, false, false, golangProxy)
}
