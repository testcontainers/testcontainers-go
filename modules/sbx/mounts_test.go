package sbx

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBindMountsRespectWorkspaceAndInnerSocket(t *testing.T) {
	workspace := t.TempDir()
	inside := filepath.Join(workspace, "inside")
	if err := os.WriteFile(inside, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	var calls []string
	m := newBindMountManager(sandboxRuntime{name: "fake", workspace: workspace, run: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, nil
	}})
	host := map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(inside+":/a") + `,` + quoteJSON(dockerSocketPath+":/var/run/docker.sock") + `]`)}
	if err := m.prepare(context.Background(), host); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 0 {
		t.Fatalf("workspace or inner socket requested host share: %v", calls)
	}
	host = map[string]json.RawMessage{"Mounts": json.RawMessage(`[{"Type":"bind","Source":` + quoteJSON(outside) + `,"Target":"/b","ReadOnly":true}]`)}
	if err := m.prepare(context.Background(), host); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || !strings.Contains(calls[0], outside+":"+outside+":ro") {
		t.Fatalf("external read-only bind not mounted: %v", calls)
	}
	if err := m.prepare(context.Background(), host); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("duplicate mount: %v", calls)
	}
}

func TestMissingOrDeniedBindFailsBeforeCreate(t *testing.T) {
	var calls []string
	m := newBindMountManager(sandboxRuntime{name: "fake", run: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, errors.New("policy denied")
	}})
	missing := filepath.Join(t.TempDir(), "absent")
	host := map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(missing+":/x") + `]`)}
	if err := m.prepare(context.Background(), host); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("missing bind error: %v", err)
	}
	existing := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(existing, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	host["Binds"] = json.RawMessage(`[` + quoteJSON(existing+":/x") + `]`)
	if err := m.prepare(context.Background(), host); err == nil || !strings.Contains(err.Error(), "policy denied") {
		t.Fatalf("policy error: %v", err)
	}
	if len(calls) != 1 || !strings.Contains(calls[0], "mount ") {
		t.Fatalf("unexpected commands: %v", calls)
	}
}

func quoteJSON(value string) string { raw, _ := json.Marshal(value); return string(raw) }

func TestSymlinkBindMountsShareCanonicalHostAtOriginalGuestPath(t *testing.T) {
	workspace := t.TempDir()
	outsideA := filepath.Join(t.TempDir(), "outside-a")
	outsideB := filepath.Join(t.TempDir(), "outside-b")
	for _, path := range []string{outsideA, outsideB} {
		if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	canonical := func(path string) string {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			t.Fatal(err)
		}
		return resolved
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(outsideA, alias); err != nil {
		t.Fatal(err)
	}
	var calls []string
	m := newBindMountManager(sandboxRuntime{name: "fake", workspace: workspace, run: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, nil
	}})
	bind := func(source string) map[string]json.RawMessage {
		return map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(source+":/guest:ro") + `]`)}
	}
	if err := m.prepare(context.Background(), bind(alias)); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "mount fake "+canonical(outsideA)+":"+alias+":ro" {
		t.Fatalf("external alias did not use canonical host and original target: %v", calls)
	}
	if err := os.Remove(alias); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideB, alias); err != nil {
		t.Fatal(err)
	}
	if err := m.prepare(context.Background(), bind(alias)); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 || calls[1] != "umount fake "+canonical(outsideA)+":"+alias || calls[2] != "mount fake "+canonical(outsideB)+":"+alias+":ro" {
		t.Fatalf("retargeted alias did not refresh old mapping: %v", calls)
	}
	inside := filepath.Join(workspace, "inside")
	if err := os.WriteFile(inside, []byte("workspace"), 0o600); err != nil {
		t.Fatal(err)
	}
	intoWorkspace := filepath.Join(t.TempDir(), "into-workspace")
	if err := os.Symlink(inside, intoWorkspace); err != nil {
		t.Fatal(err)
	}
	if err := m.prepare(context.Background(), bind(intoWorkspace)); err != nil {
		t.Fatal(err)
	}
	if got := calls[len(calls)-1]; got != "mount fake "+canonical(inside)+":"+intoWorkspace+":ro" {
		t.Fatalf("external alias into workspace was skipped: %v", calls)
	}
	fromWorkspace := filepath.Join(workspace, "outward")
	if err := os.Symlink(outsideA, fromWorkspace); err != nil {
		t.Fatal(err)
	}
	if err := m.prepare(context.Background(), bind(fromWorkspace)); err != nil {
		t.Fatal(err)
	}
	if got := calls[len(calls)-1]; got != "mount fake "+canonical(outsideA)+":"+fromWorkspace+":ro" {
		t.Fatalf("workspace alias to external source was skipped: %v", calls)
	}
	internalAlias := filepath.Join(workspace, "internal")
	if err := os.Symlink(inside, internalAlias); err != nil {
		t.Fatal(err)
	}
	before := len(calls)
	if err := m.prepare(context.Background(), bind(internalAlias)); err != nil {
		t.Fatal(err)
	}
	if len(calls) != before {
		t.Fatalf("internal workspace alias created unnecessary share: %v", calls)
	}
}

func TestWindowsBindPathsRejectBeforeMountAndNamedVolumeStillAllowed(t *testing.T) {
	var calls []string
	m := newBindMountManager(sandboxRuntime{name: "fake", run: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, nil
	}})
	for _, bind := range []string{`C:\data:/guest`, `C:/data:/guest`} {
		t.Run(bind, func(t *testing.T) {
			host := map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(bind) + `]`)}
			if err := m.prepare(context.Background(), host); err == nil || !strings.Contains(strings.ToLower(err.Error()), "windows") {
				t.Fatalf("Windows bind was not explicitly rejected: %v", err)
			}
			if len(calls) != 0 {
				t.Fatalf("unsupported bind caused mount side effect: %v", calls)
			}
		})
	}
	host := map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON("named-volume:/guest") + `]`)}
	if err := m.prepare(context.Background(), host); err != nil {
		t.Fatalf("named volume rejected: %v", err)
	}
	if len(calls) != 0 {
		t.Fatalf("named volume requested host mount: %v", calls)
	}
}
