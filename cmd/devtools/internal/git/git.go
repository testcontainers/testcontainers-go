package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
)

type GitClient struct {
	ctx           context.Context
	defaultBranch string
	dryRun        bool
	origin        string
}

func New(ctx context.Context, branch string, dryRun bool) *GitClient {
	if branch == "" {
		branch = "main"
	}

	return &GitClient{
		ctx:           ctx,
		defaultBranch: branch,
		dryRun:        dryRun,
		origin:        "git@github.com:testcontainers/testcontainers-go.git",
	}
}

// InitRepository initializes a git repository in the root directory of the context.
// Handy for testing.
func (g *GitClient) InitRepository() error {
	// set a fake origin for testing purposes
	g.origin = "git@testing-github.com:testcontainers/testcontainers-go.git"

	if err := g.Exec("init"); err != nil {
		return err
	}

	// URL is not real, just for testing purposes, but the name must be origin
	if err := g.Exec("remote", "add", "origin", g.origin); err != nil {
		return err
	}

	if err := g.Exec("checkout", "-b", g.defaultBranch); err != nil {
		return err
	}

	keepFile := filepath.Join(g.ctx.RootDir, ".keep")
	if _, err := os.Create(keepFile); err != nil {
		return fmt.Errorf("error creating .keep file: %w", err)
	}

	if err := g.Exec("add", ".keep"); err != nil {
		return err
	}

	if err := g.Exec("commit", "-m", "'Initial commit'"); err != nil {
		return err
	}

	return nil
}

func (g *GitClient) Exec(args ...string) error {
	if g.dryRun {
		fmt.Printf("Executing 'git %s'\n", strings.Join(args, " "))
		return nil
	}

	bashArgs := []string{"-c", "git " + strings.Join(args, " ")}

	cmd := exec.Command("/bin/bash", bashArgs...)
	cmd.Dir = g.ctx.RootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bash '%s' failed: %w", bashArgs, err)
	}

	return nil
}

func (g *GitClient) Add(files ...string) error {
	args := append([]string{"add"}, files...)

	return g.Exec(args...)
}

func (g *GitClient) Commit(msg string) error {
	return g.Exec("commit", "-m", "'"+msg+"'")
}

func (g *GitClient) ListTags() (string, error) {
	args := []string{
		"tag", "--list", "--sort=-v:refname",
	}

	return g.ExecWithOutput(args...)
}

func (g *GitClient) Log() (string, error) {
	args := []string{
		"log", "--color", "--graph", `--pretty=format:'%h -%d %s'`, "--abbrev-commit",
	}

	return g.ExecWithOutput(args...)
}

func (g *GitClient) ExecWithOutput(args ...string) (string, error) {
	if g.dryRun {
		fmt.Printf("Executing 'git %s'\n", strings.Join(args, " "))
		return "", nil
	}

	bashArgs := []string{"-c", "git " + strings.Join(args, " ")}

	cmd := exec.Command("/bin/bash", bashArgs...)
	cmd.Dir = g.ctx.RootDir

	var outbuf, errbuf strings.Builder
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	if err := cmd.Run(); err != nil {
		return errbuf.String(), fmt.Errorf("bash '%s' failed: %w", bashArgs, err)
	}

	return outbuf.String(), nil
}

func (g *GitClient) PushTags() error {
	return g.Exec("push", "origin", "--tags")
}

func (g *GitClient) remotes() (map[string]string, error) {
	args := []string{
		"remote", "-v",
	}

	lines, err := g.ExecWithOutput(args...)
	if err != nil {
		return nil, err
	}

	remotes := make(map[string]string)
	for _, line := range strings.Split(lines, "\n") {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if strings.EqualFold(parts[2], "(push)") || strings.EqualFold(parts[2], "(fetch)") {
			remotes[parts[0]+"-"+parts[2]] = parts[1]
		}
	}

	return remotes, nil
}

func (g *GitClient) Tag(tag string) error {
	err := g.Exec("tag", "-d", tag)
	if err != nil {
		// do not fail if tag does not exist
		fmt.Println("Error deleting tag", err)
	}

	return g.Exec("tag", tag)
}

// HasOriginRemote checks if the repository has an origin remote set to the expected value.
// The expected value is set in the origin field of the GitClient,
// and defaults to the testcontainers-go upstream repository.
func (g *GitClient) HasOriginRemote() error {
	remotes, err := g.remotes()
	if err != nil {
		return err
	}

	if g.dryRun {
		// no errors in dry run
		return nil
	}

	if len(remotes) == 0 {
		return fmt.Errorf("no remotes found")
	}

	// verify the origin remote exists
	var origin string
	if orp, ok := remotes["origin-(push)"]; ok {
		origin = orp
	} else if orf, ok := remotes["origin-(fetch)"]; ok {
		origin = orf
	} else {
		return fmt.Errorf("origin remote not found")
	}

	if origin != g.origin {
		return fmt.Errorf("origin remote, %s, is not the expected one: %s", origin, g.origin)
	}

	return nil
}
