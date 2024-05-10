package release

import (
	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
)

const (
	repository      = "github.com/testcontainers/testcontainers-go"
	nonReleasedText = `Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>`
)

var directories = []string{"examples", "modules"}

type Releaser interface {
	PreRun(ctx context.Context) error
	Run(ctx context.Context) error
}

type dryRunReleaseManager struct {
	releaseManager
}

func NewReleaseManager(branch string, dryRun bool) Releaser {
	if dryRun {
		return &dryRunReleaseManager{
			releaseManager: releaseManager{
				branch: branch,
			},
		}
	}

	return &releaseManager{
		branch: branch,
	}
}

func (p *dryRunReleaseManager) PreRun(ctx context.Context) error {
	return preRun(ctx, p.branch, true)
}

func (p *dryRunReleaseManager) Run(ctx context.Context) error {
	return run(ctx, p.branch, true)
}

type releaseManager struct {
	branch string
}

func (p *releaseManager) PreRun(ctx context.Context) error {
	return preRun(ctx, p.branch, false)
}

func (p *releaseManager) Run(ctx context.Context) error {
	return run(ctx, p.branch, false)
}
