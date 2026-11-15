package sbx

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	baseImage    = "docker.io/docker/sbx-kit-shell:1.0.0"
	dockerTCPKit = "sbx/docker-tcp:latest"
)

type configuration struct {
	binary    string
	workspace string
	timeout   time.Duration
	appName   string
	bindRoots []configuredBindRoot
}

func readConfiguration(opts options) (configuration, error) {
	cfg := configuration{binary: os.Getenv("TESTCONTAINERS_SBX_BINARY"), workspace: os.Getenv("TESTCONTAINERS_SBX_WORKSPACE"), timeout: 5 * time.Minute, appName: os.Getenv("TESTCONTAINERS_SBX_APP_NAME")}
	if opts.binary != "" {
		cfg.binary = opts.binary
	}
	if opts.workspace != "" {
		cfg.workspace = opts.workspace
	}
	if opts.appName != nil {
		cfg.appName = *opts.appName
	}
	if cfg.binary == "" {
		cfg.binary = "sbx"
	}
	if cfg.workspace == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return cfg, fmt.Errorf("get sbx workspace: %w", err)
		}
		cfg.workspace = cwd
	}
	workspace, err := filepath.Abs(cfg.workspace)
	if err != nil {
		return cfg, fmt.Errorf("resolve sbx workspace: %w", err)
	}
	cfg.workspace = filepath.Clean(workspace)
	workspaceRoot, err := configureBindRoot(cfg.workspace, true)
	if err != nil {
		return cfg, fmt.Errorf("configure sbx workspace: %w", err)
	}
	cfg.bindRoots = append(cfg.bindRoots, workspaceRoot)
	bindRoots := opts.bindRoots
	if bindRoots == nil {
		if raw := os.Getenv("TESTCONTAINERS_SBX_BIND_ROOTS"); raw != "" {
			bindRoots = strings.Split(raw, string(os.PathListSeparator))
		}
	}
	for _, root := range bindRoots {
		configured, configureErr := configureBindRoot(root, false)
		if configureErr != nil {
			return cfg, fmt.Errorf("configure sbx bind root %q: %w", root, configureErr)
		}
		cfg.bindRoots = append(cfg.bindRoots, configured)
	}
	if opts.timeout > 0 {
		cfg.timeout = opts.timeout
	} else if raw := os.Getenv("TESTCONTAINERS_SBX_TIMEOUT"); raw != "" {
		cfg.timeout, err = time.ParseDuration(raw)
		if err != nil || cfg.timeout <= 0 {
			return cfg, fmt.Errorf("TESTCONTAINERS_SBX_TIMEOUT must be a positive Go duration: %q", raw)
		}
	}
	return cfg, nil
}

func configureBindRoot(path string, workspace bool) (configuredBindRoot, error) {
	if !filepath.IsAbs(path) {
		return configuredBindRoot{}, errors.New("path must be absolute")
	}
	logical := filepath.Clean(path)
	canonical, err := filepath.EvalSymlinks(logical)
	if err != nil {
		return configuredBindRoot{}, fmt.Errorf("resolve path: %w", err)
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return configuredBindRoot{}, fmt.Errorf("make resolved path absolute: %w", err)
	}
	root, err := os.OpenRoot(canonical)
	if err != nil {
		return configuredBindRoot{}, fmt.Errorf("open directory: %w", err)
	}
	if err := root.Close(); err != nil {
		return configuredBindRoot{}, fmt.Errorf("close directory: %w", err)
	}
	return configuredBindRoot{path: logical, canonical: filepath.Clean(canonical), workspace: workspace}, nil
}
