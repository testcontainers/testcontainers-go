package sbx

import (
	"fmt"
	"os"
	"path/filepath"
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
}

func readConfiguration() (configuration, error) {
	cfg := configuration{binary: os.Getenv("TESTCONTAINERS_SBX_BINARY"), workspace: os.Getenv("TESTCONTAINERS_SBX_WORKSPACE"), timeout: 5 * time.Minute, appName: os.Getenv("TESTCONTAINERS_SBX_APP_NAME")}
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
	cfg.workspace = workspace
	if raw := os.Getenv("TESTCONTAINERS_SBX_TIMEOUT"); raw != "" {
		cfg.timeout, err = time.ParseDuration(raw)
		if err != nil || cfg.timeout <= 0 {
			return cfg, fmt.Errorf("TESTCONTAINERS_SBX_TIMEOUT must be a positive Go duration: %q", raw)
		}
	}
	return cfg, nil
}
