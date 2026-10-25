package sbx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestBindMountsRespectWorkspaceAndInnerSocket(t *testing.T) {
	workspace := canonicalTempDir(t)
	inside := filepath.Join(workspace, "inside")
	if err := os.WriteFile(inside, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	outsideRoot := canonicalTempDir(t)
	outside := filepath.Join(outsideRoot, "outside")
	if err := os.WriteFile(outside, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	var calls []string
	m := newTestBindMountManager(t, sandboxRuntime{name: "fake", run: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, nil
	}}, workspace, outsideRoot)
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
	host = map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(outside+":/b:rw") + `]`)}
	if err := m.prepare(context.Background(), host); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 || calls[1] != "umount fake "+outside+":"+outside || calls[2] != "mount fake "+outside+":"+outside {
		t.Fatalf("read-write bind did not refresh read-only outer mount: %v", calls)
	}
}

func TestMissingOrDeniedBindFailsBeforeCreate(t *testing.T) {
	var calls []string
	workspace := t.TempDir()
	allowed := t.TempDir()
	m := newTestBindMountManager(t, sandboxRuntime{name: "fake", run: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, errors.New("policy denied")
	}}, workspace, allowed)
	missing := filepath.Join(allowed, "absent")
	host := map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(missing+":/x") + `]`)}
	if err := m.prepare(context.Background(), host); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("missing bind error: %v", err)
	}
	unapproved := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(unapproved, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	host["Binds"] = json.RawMessage(`[` + quoteJSON(unapproved+":/x") + `]`)
	if err := m.prepare(context.Background(), host); err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("unapproved bind error: %v", err)
	}
	sibling := allowed + "-sibling"
	if err := os.Mkdir(sibling, 0o700); err != nil {
		t.Fatal(err)
	}
	siblingFile := filepath.Join(sibling, "file")
	if err := os.WriteFile(siblingFile, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	tricks := []struct {
		name   string
		source string
		want   string
	}{
		{name: "sibling prefix", source: siblingFile, want: "outside"},
		{name: "dot-dot escape", source: allowed + string(filepath.Separator) + ".." + string(filepath.Separator) + filepath.Base(sibling) + string(filepath.Separator) + "file", want: "must not contain"},
	}
	for _, trick := range tricks {
		t.Run(trick.name, func(t *testing.T) {
			host["Binds"] = json.RawMessage(`[` + quoteJSON(trick.source+":/x") + `]`)
			if err := m.prepare(context.Background(), host); err == nil || !strings.Contains(err.Error(), trick.want) {
				t.Fatalf("path trick %q accepted: %v", trick.source, err)
			}
			if len(calls) != 0 {
				t.Fatalf("path trick %q caused a mount side effect: %v", trick.source, calls)
			}
		})
	}
	existing := filepath.Join(allowed, "file")
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
	workspace := canonicalTempDir(t)
	external := canonicalTempDir(t)
	outsideA := filepath.Join(external, "outside-a")
	outsideB := filepath.Join(external, "outside-b")
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
	alias := filepath.Join(external, "alias")
	if err := os.Symlink(filepath.Base(outsideA), alias); err != nil {
		t.Fatal(err)
	}
	var calls []string
	m := newTestBindMountManager(t, sandboxRuntime{name: "fake", run: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, nil
	}}, workspace, external)
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
	if err := os.Symlink(filepath.Base(outsideB), alias); err != nil {
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
	internalAlias := filepath.Join(workspace, "internal")
	if err := os.Symlink(filepath.Base(inside), internalAlias); err != nil {
		t.Fatal(err)
	}
	before := len(calls)
	if err := m.prepare(context.Background(), bind(internalAlias)); err != nil {
		t.Fatal(err)
	}
	if len(calls) != before {
		t.Fatalf("internal workspace alias created unnecessary share: %v", calls)
	}
	deep := filepath.Join(external, "deep")
	if err := os.MkdirAll(filepath.Join(deep, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	deepFile := filepath.Join(deep, "file")
	if err := os.WriteFile(deepFile, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	nestedLink := filepath.Join(external, "nested-link")
	if err := os.Symlink(filepath.Join("deep", "nested"), nestedLink); err != nil {
		t.Fatal(err)
	}
	parentTarget := filepath.Join(external, "parent-target")
	parentLinkTarget := "nested-link" + string(filepath.Separator) + ".." + string(filepath.Separator) + "file"
	if err := os.Symlink(parentLinkTarget, parentTarget); err != nil {
		t.Fatal(err)
	}
	before = len(calls)
	if err := m.prepare(context.Background(), bind(parentTarget)); err != nil {
		t.Fatalf("confined parent component in symlink target rejected: %v", err)
	}
	if len(calls) != before+1 || calls[before] != "mount fake "+deepFile+":"+parentTarget+":ro" {
		t.Fatalf("symlink target was cleaned before filesystem resolution: %v", calls[before:])
	}

	absolute := filepath.Join(external, "absolute")
	if err := os.Symlink(outsideA, absolute); err != nil {
		t.Fatal(err)
	}
	if err := m.prepare(context.Background(), bind(absolute)); err == nil || !strings.Contains(err.Error(), "absolute symbolic") {
		t.Fatalf("absolute symlink accepted: %v", err)
	}
	escape := filepath.Join(external, "escape")
	if err := os.Symlink(filepath.Join("..", filepath.Base(workspace), "inside"), escape); err != nil {
		t.Fatal(err)
	}
	if err := m.prepare(context.Background(), bind(escape)); err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("relative symlink escape accepted: %v", err)
	}
	chainTarget := filepath.Base(outsideA)
	for i := 8; i >= 0; i-- {
		name := fmt.Sprintf("chain-%d", i)
		if err := os.Symlink(chainTarget, filepath.Join(external, name)); err != nil {
			t.Fatal(err)
		}
		chainTarget = name
	}
	chain := filepath.Join(external, chainTarget)
	before = len(calls)
	if err := m.prepare(context.Background(), bind(chain)); err != nil {
		t.Fatalf("nine-link in-root chain rejected: %v", err)
	}
	if len(calls) != before+1 || calls[before] != "mount fake "+canonical(outsideA)+":"+chain+":ro" {
		t.Fatalf("nine-link chain did not resolve to its in-root target: %v", calls[before:])
	}
}

func TestWindowsBindPathsRejectBeforeMountAndNamedVolumeStillAllowed(t *testing.T) {
	var calls []string
	m := newTestBindMountManager(t, sandboxRuntime{name: "fake", run: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, nil
	}}, t.TempDir())
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

func TestBindRootUnionAllowsSymlinkWithinBroaderRoot(t *testing.T) {
	workspace := canonicalTempDir(t)
	broad := canonicalTempDir(t)
	narrow := filepath.Join(broad, "narrow")
	sibling := filepath.Join(broad, "sibling")
	if err := os.MkdirAll(narrow, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sibling, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sibling, "data")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(narrow, "alias")
	if err := os.Symlink(filepath.Join("..", "sibling", "data"), alias); err != nil {
		t.Fatal(err)
	}
	var calls []string
	m := newTestBindMountManager(t, sandboxRuntime{name: "fake", run: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, nil
	}}, workspace, broad, narrow)
	host := map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(alias+":/data:ro") + `]`)}
	if err := m.prepare(context.Background(), host); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "mount fake "+target+":"+alias+":ro" {
		t.Fatalf("broader authorized root did not accept nested symlink: %v", calls)
	}
}

func TestConfiguredRootAliasPreservesGuestPath(t *testing.T) {
	parent := canonicalTempDir(t)
	realWorkspace := filepath.Join(parent, "real-workspace")
	if err := os.Mkdir(realWorkspace, 0o700); err != nil {
		t.Fatal(err)
	}
	logicalWorkspace := filepath.Join(parent, "logical-workspace")
	if err := os.Symlink(filepath.Base(realWorkspace), logicalWorkspace); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(realWorkspace, "data")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	configured, err := configureBindRoot(logicalWorkspace, true)
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	m, err := newBindMountManager(sandboxRuntime{
		name:      "fake",
		workspace: logicalWorkspace,
		bindRoots: []configuredBindRoot{configured},
		run: func(_ context.Context, args ...string) ([]byte, error) {
			calls = append(calls, strings.Join(args, " "))
			return nil, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.close() })

	logicalSource := filepath.Join(logicalWorkspace, "data")
	host := map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(logicalSource+":/data:ro") + `]`)}
	if err := m.prepare(context.Background(), host); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 0 {
		t.Fatalf("logical workspace path created an unnecessary mount: %v", calls)
	}
	host["Binds"] = json.RawMessage(`[` + quoteJSON(file+":/canonical:ro") + `]`)
	if err := m.prepare(context.Background(), host); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "mount fake "+file+":"+file+":ro" {
		t.Fatalf("canonical workspace alias was not mounted at its requested guest path: %v", calls)
	}
}

func TestWorkspaceReplacementRejectedBeforeCreate(t *testing.T) {
	for _, alias := range []bool{false, true} {
		t.Run(fmt.Sprintf("symlink-%t", alias), func(t *testing.T) {
			parent := canonicalTempDir(t)
			workspace := filepath.Join(parent, "workspace")
			if err := os.Mkdir(workspace, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(workspace, "data"), []byte("old"), 0o600); err != nil {
				t.Fatal(err)
			}
			logical := workspace
			if alias {
				logical = filepath.Join(parent, "alias")
				if err := os.Symlink(filepath.Base(workspace), logical); err != nil {
					t.Fatal(err)
				}
			}
			m := newTestBindMountManager(t, sandboxRuntime{name: "fake", run: func(_ context.Context, args ...string) ([]byte, error) {
				t.Errorf("replaced workspace caused a mount side effect: %v", args)
				return nil, nil
			}}, logical)
			if err := os.Rename(workspace, workspace+"-old"); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(workspace, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(workspace, "data"), []byte("replacement"), 0o600); err != nil {
				t.Fatal(err)
			}
			host := bindRequestForTest(filepath.Join(logical, "data") + ":/data")
			for _, check := range []func() error{
				func() error { return m.validate(host) },
				func() error { return m.prepare(context.Background(), host) },
			} {
				if err := check(); err == nil || !strings.Contains(err.Error(), "changed after initialization") {
					t.Fatalf("replaced workspace accepted: %v", err)
				}
			}
		})
	}
}

func TestBindPreparationRollsBackBatch(t *testing.T) {
	for _, mode := range []string{"rollback", "cancelled", "cleanup-retry", "restore-read-only", "restore-after-mount-failure", "replaced-source"} {
		t.Run(mode, func(t *testing.T) {
			external := canonicalTempDir(t)
			existing := filepath.Join(external, "a-existing")
			first := filepath.Join(external, "b-first")
			last := filepath.Join(external, "c-last")
			for _, path := range []string{existing, first, last} {
				if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			mountErr := errors.New("mount denied")
			cleanupErr := errors.New("umount unavailable")
			active := make(map[string]string)
			var calls []string
			cleanupFailed := false
			m := newTestBindMountManager(t, sandboxRuntime{name: "fake", run: func(ctx context.Context, args ...string) ([]byte, error) {
				calls = append(calls, strings.Join(args, " "))
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				spec := args[2]
				key := strings.TrimSuffix(spec, ":ro")
				if args[0] == "umount" {
					if mode == "cleanup-retry" && strings.HasPrefix(spec, first+":") && !cleanupFailed {
						cleanupFailed = true
						return nil, cleanupErr
					}
					delete(active, key)
					return nil, nil
				}
				if strings.HasPrefix(spec, last+":") || (mode == "restore-after-mount-failure" && spec == first+":"+first) {
					if mode == "cancelled" {
						cancel()
					}
					return nil, mountErr
				}
				active[key] = spec
				return nil, nil
			}}, canonicalTempDir(t), external)
			if err := m.prepare(ctx, bindRequestForTest(existing+":/existing:ro")); err != nil {
				t.Fatal(err)
			}
			restore := strings.HasPrefix(mode, "restore-")
			if restore || mode == "replaced-source" {
				if err := m.prepare(ctx, bindRequestForTest(first+":/first:ro")); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "replaced-source" {
				if err := os.Rename(first, first+"-old"); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(first, []byte("replacement"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			err := m.prepare(ctx, bindRequestForTest(existing+":/existing:ro", first+":/first", last+":/last"))
			if !errors.Is(err, mountErr) {
				t.Fatalf("lost preparation error: %v", err)
			}
			if mode == "cleanup-retry" {
				if !errors.Is(err, cleanupErr) || active[first+":"+first] == "" {
					t.Fatalf("lost failed cleanup: active=%v err=%v", active, err)
				}
				// Even a create without binds must reconcile a failed rollback.
				if err := m.prepare(context.Background(), nil); err != nil {
					t.Fatal(err)
				}
			}
			if active[existing+":"+existing] != existing+":"+existing+":ro" {
				t.Fatalf("rollback removed a shared mount: %v", active)
			}
			if restore {
				if len(active) != 2 || active[first+":"+first] != first+":"+first+":ro" {
					t.Fatalf("rollback did not restore previous permissions: %v; calls=%v", active, calls)
				}
			} else if len(active) != 1 || len(m.mounted) != 1 {
				t.Fatalf("failed batch retained mounts: active=%v cache=%v calls=%v", active, m.mounted, calls)
			}
			if mode == "replaced-source" {
				if err := m.prepare(context.Background(), bindRequestForTest(first+":/first")); err != nil {
					t.Fatalf("stale mount prevented a fresh request: %v", err)
				}
			}
		})
	}
}

func bindRequestForTest(binds ...string) map[string]json.RawMessage {
	raw, _ := json.Marshal(binds)
	return map[string]json.RawMessage{"Binds": raw}
}

func TestBindRootReplacementRejectedBeforeMount(t *testing.T) {
	workspace := t.TempDir()
	externalParent := t.TempDir()
	external := filepath.Join(externalParent, "allowed")
	if err := os.Mkdir(external, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(external, "data")
	if err := os.WriteFile(source, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	var calls []string
	m := newTestBindMountManager(t, sandboxRuntime{name: "fake", run: func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		return nil, nil
	}}, workspace, external)
	if err := os.Rename(external, external+"-old"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(external, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	host := map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(source+":/data") + `]`)}
	if err := m.prepare(context.Background(), host); err == nil || !strings.Contains(err.Error(), "changed after initialization") {
		t.Fatalf("replaced configured root accepted: %v", err)
	}
	if len(calls) != 0 {
		t.Fatalf("replaced root caused a mount side effect: %v", calls)
	}
}

func TestSpecialBindSourceAndConcurrentClose(t *testing.T) {
	workspace := t.TempDir()
	external := t.TempDir()
	socket := filepath.Join(external, "service.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	m := newTestBindMountManager(t, sandboxRuntime{name: "fake", run: func(_ context.Context, _ ...string) ([]byte, error) {
		once.Do(func() { close(started) })
		<-release
		return nil, nil
	}}, workspace, external)
	socketHost := map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(socket+":/socket") + `]`)}
	if err := m.prepare(context.Background(), socketHost); err == nil || !strings.Contains(err.Error(), "host socket") {
		t.Fatalf("host socket accepted: %v", err)
	}

	file := filepath.Join(external, "data")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	fileHost := map[string]json.RawMessage{"Binds": json.RawMessage(`[` + quoteJSON(file+":/data") + `]`)}
	prepareDone := make(chan error, 1)
	go func() { prepareDone <- m.prepare(context.Background(), fileHost) }()
	<-started
	closeDone := make(chan error, 1)
	go func() { closeDone <- m.close() }()
	close(release)
	if err := <-prepareDone; err != nil {
		t.Fatal(err)
	}
	if err := <-closeDone; err != nil {
		t.Fatal(err)
	}
	if err := m.close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
	if err := m.prepare(context.Background(), fileHost); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("closed manager accepted prepare: %v", err)
	}
}

func newTestBindMountManager(t *testing.T, runtime sandboxRuntime, workspace string, external ...string) *bindMountManager {
	t.Helper()
	runtime.workspace = workspace
	paths := append([]string{workspace}, external...)
	for i, path := range paths {
		root, err := configureBindRoot(path, i == 0)
		if err != nil {
			t.Fatal(err)
		}
		runtime.bindRoots = append(runtime.bindRoots, root)
	}
	m, err := newBindMountManager(runtime)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.close(); err != nil {
			t.Error(err)
		}
	})
	return m
}
