package sbx

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigurationDefaultsAndEnvironment(t *testing.T) {
	for _, key := range []string{"BINARY", "WORKSPACE", "TIMEOUT", "APP_NAME", "BIND_ROOTS"} {
		t.Setenv("TESTCONTAINERS_SBX_"+key, "")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := readConfiguration(options{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.binary != "sbx" || cfg.workspace != cwd || cfg.timeout != 5*time.Minute || cfg.appName != "" || len(cfg.bindRoots) != 1 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	workspace, root := canonicalTempDir(t), canonicalTempDir(t)
	t.Setenv("TESTCONTAINERS_SBX_BINARY", "/custom/sbx")
	t.Setenv("TESTCONTAINERS_SBX_WORKSPACE", workspace)
	t.Setenv("TESTCONTAINERS_SBX_TIMEOUT", "8m")
	t.Setenv("TESTCONTAINERS_SBX_APP_NAME", "from-env")
	t.Setenv("TESTCONTAINERS_SBX_BIND_ROOTS", root)
	var p Provider
	if err := p.Configure(WithKits("./tools")); err != nil {
		t.Fatal(err)
	}
	cfg, err = readConfiguration(p.options.clone())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.binary != "/custom/sbx" || cfg.workspace != workspace || cfg.timeout != 8*time.Minute || cfg.appName != "from-env" || len(cfg.bindRoots) != 2 || cfg.bindRoots[1].path != root {
		t.Fatalf("kit options discarded environment settings: %+v", cfg)
	}
}

func TestConfigurationOptionPrecedence(t *testing.T) {
	workspace, root := canonicalTempDir(t), canonicalTempDir(t)
	t.Setenv("TESTCONTAINERS_SBX_BINARY", "from-env")
	t.Setenv("TESTCONTAINERS_SBX_WORKSPACE", filepath.Join(workspace, "missing"))
	t.Setenv("TESTCONTAINERS_SBX_TIMEOUT", "invalid")
	t.Setenv("TESTCONTAINERS_SBX_APP_NAME", "from-env")
	t.Setenv("TESTCONTAINERS_SBX_BIND_ROOTS", "relative")
	paths := []string{root}
	binds := WithBindRoots(paths...)
	paths[0] = "/changed"
	var p Provider
	if err := p.Configure(WithBinary("first"), WithWorkspace("./first"), WithStartupTimeout(time.Second), WithAppName("first"), WithBindRoots("/first")); err != nil {
		t.Fatal(err)
	}
	if err := p.Configure(WithBinary("custom-sbx"), WithWorkspace(workspace), WithStartupTimeout(8*time.Minute), WithAppName("test-suite"), binds); err != nil {
		t.Fatal(err)
	}
	cfg, err := readConfiguration(p.options.clone())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.binary != "custom-sbx" || cfg.workspace != workspace || cfg.timeout != 8*time.Minute || cfg.appName != "test-suite" {
		t.Fatalf("options did not override previous settings and environment: %+v", cfg)
	}
	if len(cfg.bindRoots) != 2 || !cfg.bindRoots[0].workspace || cfg.bindRoots[1].path != root {
		t.Fatalf("bind roots were not copied and replaced: %+v", cfg.bindRoots)
	}
	if err := p.Configure(WithBindRoots(), WithAppName("")); err != nil {
		t.Fatal(err)
	}
	cfg, err = readConfiguration(p.options.clone())
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.bindRoots) != 1 || !cfg.bindRoots[0].workspace || cfg.appName != "" {
		t.Fatalf("empty options did not clear inherited settings: %+v", cfg)
	}
	if os.Getenv("TESTCONTAINERS_SBX_APP_NAME") != "from-env" || os.Getenv("TESTCONTAINERS_SBX_TIMEOUT") != "invalid" {
		t.Fatal("options mutated the process environment")
	}
}

func TestConfigurationValidatesPathsAtStartup(t *testing.T) {
	t.Setenv("TESTCONTAINERS_SBX_BIND_ROOTS", "")
	t.Setenv("TESTCONTAINERS_SBX_TIMEOUT", "")
	base := canonicalTempDir(t)
	workspace := filepath.Join(base, "workspace")
	root := filepath.Join(base, "fixtures")
	var p Provider
	if err := p.Configure(WithWorkspace(workspace), WithBindRoots(root)); err != nil {
		t.Fatalf("configuration eagerly required directories: %v", err)
	}
	if _, err := readConfiguration(p.options); err == nil {
		t.Fatal("missing workspace accepted at startup")
	}
	if err := os.Mkdir(workspace, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := readConfiguration(p.options); err == nil {
		t.Fatal("missing bind root accepted at startup")
	}
	if err := os.WriteFile(root, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readConfiguration(p.options); err == nil {
		t.Fatal("nondirectory bind root accepted at startup")
	}
	if err := os.Remove(root); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Chdir(base)
	if err := p.Configure(WithWorkspace("./workspace")); err != nil {
		t.Fatal(err)
	}
	cfg, err := readConfiguration(p.options.clone())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.workspace != workspace || cfg.bindRoots[1].path != root {
		t.Fatalf("paths not resolved at startup: %+v", cfg)
	}
}
