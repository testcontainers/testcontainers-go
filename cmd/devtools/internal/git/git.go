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
	ctx    context.Context
	dryRun bool
}

func New(ctx context.Context, dryRun bool) *GitClient {
	return &GitClient{
		ctx:    ctx,
		dryRun: dryRun,
	}
}

// InitRepository initializes a git repository in the root directory of the context.
// Handy for testing.
func (g *GitClient) InitRepository() error {
	if g.dryRun {
		fmt.Println("git init")
		fmt.Println("touch .keep")
		fmt.Println("git add .keep")
		fmt.Println("git commit -m 'Initial commit'")
		return nil
	}

	if err := g.Exec("init"); err != nil {
		return err
	}

	keepFile := filepath.Join(g.ctx.RootDir, ".keep")
	if _, err := os.Create(keepFile); err != nil {
		return fmt.Errorf("error creating .keep file: %w", err)
	}

	if err := g.Exec("add", ".keep"); err != nil {
		return err
	}

	if err := g.Exec("commit", "-m", "Initial commit"); err != nil {
		return err
	}

	return nil
}

func (g *GitClient) Exec(args ...string) error {
	if g.dryRun {
		fmt.Printf("git %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = g.ctx.RootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git command failed: %w", err)
	}

	return nil
}
