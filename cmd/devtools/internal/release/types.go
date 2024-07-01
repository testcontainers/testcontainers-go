package release

import (
	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
)

const (
	golangProxy     string = "https://proxy.golang.org"
	nonReleasedText string = `Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>`
)

var directories = []string{"examples", "modules"}

type Releaser interface {
	PreRun(ctx context.Context) error
	Run(ctx context.Context) error
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

func (p *dryRunReleaseManager) PreRun(ctx context.Context) error {
	return preRun(ctx, p.branch, true)
}

func (p *dryRunReleaseManager) Run(ctx context.Context) error {
	// dry run and skip remote operations
	return run(ctx, p.branch, p.bumpType, true, true, golangProxy)
}

type releaseManager struct {
	branch   string
	bumpType string
}

func (p *releaseManager) PreRun(ctx context.Context) error {
	return preRun(ctx, p.branch, false)
}

func (p *releaseManager) Run(ctx context.Context) error {
	// no dry run and execute remote operations
	return run(ctx, p.branch, p.bumpType, false, false, golangProxy)
}
