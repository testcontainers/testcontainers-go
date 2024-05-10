package release

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	devcontext "github.com/testcontainers/testcontainers-go/devtools/internal/context"
	"github.com/testcontainers/testcontainers-go/devtools/internal/git"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	repository        = "github.com/testcontainers/testcontainers-go"
	semverDockerImage = "docker.io/mdelapenya/semver-tool:3.4.0"
)

// run performs the release
func run(ctx devcontext.Context, branch string, bumpType string, dryRun bool, skipRemoteOps bool) error {
	if bumpType == "" {
		bumpType = "minor"
	}

	version, err := extractCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to extract the current version: %w", err)
	}

	fmt.Println("Current version:", version)

	releaseVersion := "v" + version

	nextVersion, err := getSemverVersion(bumpType, releaseVersion)
	if err != nil {
		return fmt.Errorf("failed to bump the version. Please check the semver-tool image and the bump type: %w", err)
	}

	fmt.Println("Next version:", nextVersion)

	// add the files to git
	gitClient := git.New(ctx, branch, dryRun)

	// the glob expressions must be quoted to avoid shell expansion
	args := [][]string{
		{"mkdocs.yml", "sonar-project.properties"},
		{"'docs/*.md'"},
		{"'docs/**/*.md'"},
		{"'examples/**/go.*'"},
		{"'modules/**/go.*'"},
	}

	for _, arg := range args {
		if err := gitClient.Add(arg...); err != nil {
			// log the error and continue to add the other files
			fmt.Printf("error adding files: %s\n", err)
			continue
		}
	}

	// Commit the project in the current state
	if err := gitClient.Commit(fmt.Sprintf("chore: use new version (%s) in modules and examples", releaseVersion)); err != nil {
		return fmt.Errorf("error committing the release: %w", err)
	}

	if err := gitClient.Tag(releaseVersion); err != nil {
		return fmt.Errorf("error tagging the release: %w", err)
	}

	for _, directory := range directories {
		path := filepath.Join(ctx.RootDir, directory)
		modules, err := getSubdirectories(path)
		if err != nil {
			return fmt.Errorf("error getting subdirectories: %w", err)
		}

		for _, module := range modules {
			moduleTag := fmt.Sprintf("%s/%s/%s", directory, module, releaseVersion) // e.g. modules/mongodb/v0.0.1
			if err := gitClient.Tag(moduleTag); err != nil {
				return fmt.Errorf("error tagging the module: %w", err)
			}
		}
	}

	fmt.Printf("Producing a %s bump of the version, from %s to %s\n", bumpType, version, nextVersion)

	if err := replaceInFile(dryRun, false, ctx.VersionFile(), version, nextVersion); err != nil {
		return fmt.Errorf("error replacing in version file: %w", err)
	}

	if err := gitClient.Add(ctx.VersionFile()); err != nil {
		return fmt.Errorf("error adding version file: %w", err)
	}

	if err := gitClient.Commit(fmt.Sprintf("chore: prepare for next %s development version cycle (%s)", bumpType, nextVersion)); err != nil {
		return fmt.Errorf("error committing the next version: %w", err)
	}

	// for testing purposes, we can skip the remote operations, like pushing the tags
	// and hitting the golang proxy to update the latest version for the core and the modules
	if !skipRemoteOps {
		if err := gitClient.PushTags(); err != nil {
			return fmt.Errorf("error pushing tags: %w", err)
		}

		if err := hitGolangProxy(dryRun, repository, releaseVersion); err != nil {
			return fmt.Errorf("error hitting the golang proxy for the core: %w", err)
		}
		for _, directory := range directories {
			path := filepath.Join(ctx.RootDir, directory)
			modules, err := getSubdirectories(path)
			if err != nil {
				return fmt.Errorf("error getting subdirectories: %w", err)
			}

			for _, module := range modules {
				modulePath := fmt.Sprintf("%s/%s", directory, module)
				if err := hitGolangProxy(dryRun, modulePath, releaseVersion); err != nil {
					return fmt.Errorf("error hitting the golang proxy for the module: %w", err)
				}
			}
		}
	}

	return nil
}

// hitGolangProxy hits the golang proxy to update the latest version
// The URL has the following format: https://proxy.golang.org/${module_path}/@v/${module_version}.info
func hitGolangProxy(dryRun bool, modulePath string, moduleVersion string) error {
	url := fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.info", modulePath, moduleVersion)

	if dryRun {
		fmt.Printf("hitting the golang proxy: %s\n", url)
		return nil
	}

	cli := http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error hitting the golang proxy: %s", resp.Status)
	}

	return nil
}

type inMemoryLogger struct {
	logs string
}

func (l *inMemoryLogger) Accept(tcl testcontainers.Log) {
	l.logs += string(tcl.Content)
}

func getSemverVersion(bumpType string, vVersion string) (string, error) {
	logger := &inMemoryLogger{}

	// get the next version from the bump type. It will use the semver-tool to bump the version
	c, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:         semverDockerImage,
			Cmd:           []string{"bump", bumpType, vVersion},
			ImagePlatform: "linux/amd64",
			LogConsumerCfg: &testcontainers.LogConsumerConfig{
				Consumers: []testcontainers.LogConsumer{logger},
			},
			WaitingFor: wait.ForExit(),
		},
		Started: true,
	})
	if err != nil {
		return "", err
	}
	defer func() {
		err := c.Terminate(context.Background())
		if err != nil {
			fmt.Println("Error terminating container", err)
		}
	}()

	// the output of the semver-tool is the new version
	newVersion := strings.TrimSpace(logger.logs)

	return newVersion, nil
}
