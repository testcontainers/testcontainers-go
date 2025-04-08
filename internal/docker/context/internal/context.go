package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const metaFile = "meta.json"

// dockerContext represents the metadata stored for a context
type dockerContext struct {
	Description string
	Fields      map[string]any
}

// endpoint represents a Docker endpoint configuration
type endpoint struct {
	Host          string `json:",omitempty"`
	SkipTLSVerify bool
}

// metadata represents a complete context configuration
type metadata struct {
	Name      string               `json:",omitempty"`
	Context   *dockerContext       `json:"metadata,omitempty"`
	Endpoints map[string]*endpoint `json:"endpoints,omitempty"`
}

// store manages Docker context metadata files
type store struct {
	root string
}

// ExtractDockerHost extracts the Docker host from the given Docker context
func ExtractDockerHost(contextName string, metaRoot string) (string, error) {
	s := &store{root: metaRoot}

	contexts, err := s.list()
	if err != nil {
		return "", fmt.Errorf("list contexts: %w", err)
	}

	for _, ctx := range contexts {
		if ctx.Name == contextName {
			ep, ok := ctx.Endpoints["docker"]
			if !ok || ep == nil || ep.Host == "" { // Check all conditions that should trigger the error
				return "", ErrDockerHostNotSet
			}
			return ep.Host, nil
		}
	}
	return "", ErrDockerHostNotSet
}

func (s *store) list() ([]*metadata, error) {
	dirs, err := s.findMetadataDirs(s.root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("find contexts: %w", err)
	}

	var contexts []*metadata
	for _, dir := range dirs {
		ctx, err := s.load(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("load context %s: %w", dir, err)
		}
		contexts = append(contexts, ctx)
	}
	return contexts, nil
}

func (s *store) load(dir string) (*metadata, error) {
	data, err := os.ReadFile(filepath.Join(dir, metaFile))
	if err != nil {
		return nil, err
	}

	var meta metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}
	return &meta, nil
}

func (s *store) findMetadataDirs(root string) ([]string, error) {
	var dirs []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if hasMetaFile(path) {
				dirs = append(dirs, path)
				return filepath.SkipDir // don't recurse into context dirs
			}
		}
		return nil
	})
	return dirs, err
}

func hasMetaFile(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, metaFile))
	return err == nil && !info.IsDir()
}
