package sbx

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed kit/docker-tcp.yaml
var dockerTCPKit []byte

func materializeKit() (string, error) {
	dir, err := os.MkdirTemp("", "tc-go-sbx-kit-")
	if err != nil {
		return "", fmt.Errorf("create sbx kit directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docker-tcp.yaml"), dockerTCPKit, 0o600); err != nil {
		_ = os.RemoveAll(dir)
		return "", fmt.Errorf("write sbx kit: %w", err)
	}
	return dir, nil
}
