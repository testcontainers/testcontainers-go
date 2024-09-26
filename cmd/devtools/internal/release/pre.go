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

// preRun prepares the release:
// - updating the version in the mkdocs.yml and sonar-project.properties files
// - updating the version in the go.mod files of the examples and modules directories
// - running the make tidy-examples and make tidy-modules commands
// - updating the version in the markdown files in the docs directory
func preRun(ctx context.Context, gitClient *git.GitClient, branch string, dryRun bool) error {
	version, err := extractCurrentVersion(ctx)
	if err != nil {
		return err
	}

	vVersion := "v" + version

	cleanUpRemote, err := gitClient.CheckOriginRemote()
	if err != nil {
		return err
	}
	defer func() {
		err := cleanUpRemote()
		if err != nil {
			fmt.Println("Error cleaning up the remote", err)
		}
	}()

	err = gitClient.Exec("checkout", branch)
	if err != nil {
		return err
	}

	err = bumpVersion(ctx, dryRun, vVersion)
	if err != nil {
		return fmt.Errorf("error bumping version: %w", err)
	}

	// loop through the examples and modules directories to update the go.mod files
	for _, directory := range directories {
		path := filepath.Join(ctx.RootDir, directory)
		modules, err := getSubdirectories(path)
		if err != nil {
			return fmt.Errorf("error getting subdirectories: %w", err)
		}

		// loop through the Go modules in the directory
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

// extractCurrentVersion extracts the current version from the version file.
// It reads the contents of that version file and replaces the value of the constant representing the version, if found.
// If the version is not found, it returns an error.
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

// bumpVersion bumps the version in the mkdocs.yml and sonar-project.properties files.
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
		fmt.Printf("Replacing '%s' with '%s' in file %s\n", old, new, filePath)
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

// processMarkdownFiles processes all the markdown files in the docs directory, replacing the non-released text with the released text.
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

func runCommand(ctx context.Context, dryRun bool, command string, arg ...string) error {
	if dryRun {
		fmt.Printf("%s %s\n", command, arg)
		return nil
	}

	cmd := exec.Command(command, arg...)
	cmd.Dir = ctx.RootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	return nil
}
