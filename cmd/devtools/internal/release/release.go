package release

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
	"github.com/testcontainers/testcontainers-go/devtools/internal/git"
)

const (
	repository      = "github.com/testcontainers/testcontainers-go"
	nonReleasedText = `Not available until the next release of testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go\"><span class=\"tc-version\">:material-tag: main</span></a>`
)

var directories = []string{"examples", "modules"}

type Releaser interface {
	PreRun(ctx context.Context) error
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

type releaseManager struct {
	branch string
}

func (p *releaseManager) PreRun(ctx context.Context) error {
	return preRun(ctx, p.branch, false)
}

func preRun(ctx context.Context, branch string, dryRun bool) error {
	version, err := extractCurrentVersion(ctx)
	if err != nil {
		return err
	}

	vVersion := "v" + version

	gitClient := git.New(ctx, branch, dryRun)

	err = gitClient.Exec("checkout", branch)
	if err != nil {
		return err
	}

	err = bumpVersion(ctx, dryRun, vVersion)
	if err != nil {
		return fmt.Errorf("error bumping version: %w", err)
	}

	for _, directory := range directories {
		path := filepath.Join(ctx.RootDir, directory)
		modules, err := getSubdirectories(path)
		if err != nil {
			return fmt.Errorf("error getting subdirectories: %w", err)
		}

		for _, module := range modules {
			moduleModFile := filepath.Join(path, module, "go.mod")
			err := replaceInFile(dryRun, true, moduleModFile, "testcontainers-go v.*", fmt.Sprintf("testcontainers-go v%s", version))
			if err != nil {
				return fmt.Errorf("error replacing in file: %w", err)
			}

		}

		// tidy all the modules in the directory
		err = runCommand(context.New(path), dryRun, "make", fmt.Sprintf("tidy-%s", directory))
		if err != nil {
			return err
		}
	}

	docsDir := filepath.Join(ctx.RootDir, "docs")
	return processMarkdownFiles(dryRun, docsDir, version)
}

func extractCurrentVersion(ctx context.Context) (string, error) {
	data, err := os.ReadFile(ctx.VersionFile())
	if err != nil {
		return "", fmt.Errorf("error reading version file: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "const Version =") {
			return strings.Trim(strings.Split(line, "\"")[1], " "), nil
		}
	}
	return "", fmt.Errorf("version not found in version.go")
}

func bumpVersion(ctx context.Context, dryRun bool, vVersion string) error {
	if err := replaceInFile(dryRun, true, ctx.MkdocsConfigFile(), "latest_version: .*", fmt.Sprintf("latest_version: %s", vVersion)); err != nil {
		return err
	}

	if err := replaceInFile(dryRun, true, ctx.SonarProjectFile(), "sonar.projectVersion=.*", fmt.Sprintf("sonar.projectVersion=%s", vVersion)); err != nil {
		return err
	}

	return nil
}

func replaceInFile(dryRun bool, regex bool, filePath string, old string, new string) error {
	if dryRun {
		fmt.Printf("sed \"s/%s/%s/g\" %s > %s.tmp\n", old, new, filePath, filePath)
		fmt.Printf("mv %s.tmp %s\n", filePath, filePath)

		return nil
	}

	read, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var newContents string
	if regex {
		r := regexp.MustCompile(old)
		newContents = r.ReplaceAllString(string(read), new)
	} else {
		newContents = strings.ReplaceAll(string(read), old, new)
	}

	err = os.WriteFile(filePath, []byte(newContents), 0)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func getSubdirectories(path string) ([]string, error) {
	var directories []string
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() && file.Name() != "_template" {
			directories = append(directories, file.Name())
		}
	}
	return directories, nil
}

func sinceVersionText(version string) string {
	return fmt.Sprintf(`Since testcontainers-go <a href=\"https://github.com/testcontainers/testcontainers-go/releases/tag/v%s\"><span class=\"tc-version\">:material-tag: v%s</span></a>`, version, version)
}

func processMarkdownFiles(dryRun bool, docsDir, version string) error {
	return filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".md") {
			releasedText := sinceVersionText(version)
			err := replaceInFile(dryRun, false, path, nonReleasedText, releasedText)
			if err != nil {
				return fmt.Errorf("error replacing in file: %w", err)
			}

			return nil
		}
		return nil
	})
}

func runCommand(ctx context.Context, dryRun bool, command string, arg string) error {
	if dryRun {
		fmt.Printf("%s %s\n", command, arg)
		return nil
	}

	cmd := exec.Command(command, arg)
	cmd.Dir = ctx.RootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	return nil
}
