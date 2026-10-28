package sbx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	code := m.Run()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := Cleanup(ctx); err != nil {
		println("sbx test cleanup:", err.Error())
		code = 1
	}
	cancel()
	os.Exit(code)
}

func TestProviderImportIsLazyAndUnusedCleanupIsSafe(t *testing.T) {
	if os.Getenv("TC_SBX_LAZY_IMPORT_CHILD") == "1" {
		if defaultProvider.started {
			t.Fatal("import initialized default sbx provider")
		}
		if err := Configure(WithKits("./unused-kit")); err != nil {
			t.Fatal(err)
		}
		if defaultProvider.started || len(defaultProvider.options.kits) != 1 {
			t.Fatal("configuration did not update the lazy default provider")
		}
		if err := Cleanup(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := (&Provider{}).Cleanup(context.Background()); err != nil {
			t.Fatal(err)
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestProviderImportIsLazyAndUnusedCleanupIsSafe$")
	command.Env = append(os.Environ(), "TC_SBX_LAZY_IMPORT_CHILD=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("isolated import check failed: %v\n%s", err, output)
	}
}

func TestConfigurationRejectsInvalidTimeout(t *testing.T) {
	for _, v := range []string{"0", "-1s", "nonsense"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("TESTCONTAINERS_SBX_TIMEOUT", v)
			if _, err := readConfiguration(); err == nil {
				t.Fatal("accepted invalid timeout")
			}
		})
	}
}

func TestConfigurationBindRoots(t *testing.T) {
	workspace := canonicalTempDir(t)
	first := canonicalTempDir(t)
	parent := canonicalTempDir(t)
	realSecond := filepath.Join(parent, "real")
	if err := os.Mkdir(realSecond, 0o700); err != nil {
		t.Fatal(err)
	}
	second := filepath.Join(parent, "alias")
	if err := os.Symlink(filepath.Base(realSecond), second); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TESTCONTAINERS_SBX_WORKSPACE", workspace)
	t.Setenv("TESTCONTAINERS_SBX_BIND_ROOTS", strings.Join([]string{first, second}, string(os.PathListSeparator)))
	cfg, err := readConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.bindRoots) != 3 {
		t.Fatalf("configured roots: %+v", cfg.bindRoots)
	}
	if !cfg.bindRoots[0].workspace || cfg.bindRoots[0].path != workspace {
		t.Fatalf("workspace root: %+v", cfg.bindRoots[0])
	}
	if cfg.bindRoots[2].path != second || cfg.bindRoots[2].canonical != realSecond {
		t.Fatalf("symlinked root did not retain logical and canonical paths: %+v", cfg.bindRoots[2])
	}
}

func TestConfigurationRejectsInvalidBindRoots(t *testing.T) {
	workspace := t.TempDir()
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TESTCONTAINERS_SBX_WORKSPACE", workspace)
	for _, value := range []string{
		"relative",
		filepath.Join(t.TempDir(), "missing"),
		file,
		workspace + string(os.PathListSeparator),
	} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("TESTCONTAINERS_SBX_BIND_ROOTS", value)
			if _, err := readConfiguration(); err == nil {
				t.Fatalf("accepted invalid bind roots %q", value)
			}
		})
	}
}

func TestDiscoverDockerURLOwnedEndpoint(t *testing.T) {
	const owned = `{"host_ip":"127.0.0.1","host_port":43219,"sandbox_port":2375,"protocol":"tcp"}`
	const owned4 = `{"host_ip":"127.0.0.1","host_port":43219,"sandbox_port":2375,"protocol":"tcp4"}`
	const kit = `{"host_ip":"127.0.0.1","host_port":43220,"sandbox_port":2375,"protocol":"tcp"},{"host_ip":"::1","host_port":43220,"sandbox_port":2375,"protocol":"tcp"}`
	const conflict = `{"host_ip":"127.0.0.1","host_port":43219,"sandbox_port":8080,"protocol":"tcp4"}`
	cliErr := errors.New("ports failed")
	cases := []struct {
		name    string
		output  string
		port    int
		runErr  error
		wantErr string
	}{
		{name: "owned first with dual stack kit", output: "[" + owned + "," + kit + "]", port: 43219},
		{name: "owned last with dual stack kit", output: "[" + kit + "," + owned + "]", port: 43219},
		{name: "tcp4", output: "[" + owned4 + "]", port: 43219},
		{name: "equivalent duplicates", output: "[" + owned + "," + owned4 + "," + owned + "]", port: 43219},
		{name: "missing owned", output: "[" + kit + "]", port: 43219, wantErr: "127.0.0.1:43219"},
		{name: "empty", output: "[]", port: 43219, wantErr: "127.0.0.1:43219"},
		{name: "wildcard only", output: `[{"host_ip":"0.0.0.0","host_port":43219,"sandbox_port":2375,"protocol":"tcp"}]`, port: 43219, wantErr: "127.0.0.1:43219"},
		{name: "IPv6 only", output: `[{"host_ip":"::1","host_port":43219,"sandbox_port":2375,"protocol":"tcp6"}]`, port: 43219, wantErr: "127.0.0.1:43219"},
		{name: "UDP only", output: `[{"host_ip":"127.0.0.1","host_port":43219,"sandbox_port":2375,"protocol":"udp"}]`, port: 43219, wantErr: "127.0.0.1:43219"},
		{name: "unrelated UDP ignored", output: "[" + owned + `,{"host_ip":"127.0.0.1","host_port":43219,"sandbox_port":8080,"protocol":"udp"}]`, port: 43219},
		{name: "unrelated IPv6 ignored", output: "[" + owned + `,{"host_ip":"::1","host_port":43219,"sandbox_port":8080,"protocol":"tcp6"}]`, port: 43219},
		{name: "tcp conflict", output: "[" + owned + `,{"host_ip":"127.0.0.1","host_port":43219,"sandbox_port":8080,"protocol":"tcp"}]`, port: 43219, wantErr: "conflicting sbx Docker mapping for 127.0.0.1:43219"},
		{name: "conflict before valid", output: "[" + conflict + "," + owned + "]", port: 43219, wantErr: "conflicting sbx Docker mapping for 127.0.0.1:43219"},
		{name: "conflict after valid", output: "[" + owned + "," + conflict + "]", port: 43219, wantErr: "conflicting sbx Docker mapping for 127.0.0.1:43219"},
		{name: "zero expected", port: 0, wantErr: "invalid expected"},
		{name: "negative expected", port: -1, wantErr: "invalid expected"},
		{name: "overflow expected", port: 65536, wantErr: "invalid expected"},
		{name: "malformed JSON", output: "not-json", port: 43219, wantErr: "decode"},
		{name: "CLI failure", port: 43219, runErr: cliErr, wantErr: "ports failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := func(_ context.Context, args ...string) ([]byte, error) {
				if got := strings.Join(args, " "); got != "ports owned --format json" {
					t.Fatalf("unexpected discovery command: %s", got)
				}
				return []byte(tc.output), tc.runErr
			}
			endpoint, err := discoverDockerURL(context.Background(), run, "owned", tc.port)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				if tc.runErr != nil && !errors.Is(err, tc.runErr) {
					t.Fatalf("lost CLI error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if endpoint.String() != "http://127.0.0.1:43219" {
				t.Fatalf("unexpected owned endpoint: %s", endpoint)
			}
		})
	}
}

func TestDockerReadinessSucceeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_ping" {
			t.Errorf("unexpected readiness path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()
	endpoint, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := waitForDocker(ctx, endpoint); err != nil {
		t.Fatal(err)
	}
}

func TestDockerReadinessHonorsDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer server.Close()
	endpoint, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := waitForDocker(ctx, endpoint); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("readiness ignored deadline")
	}
}

type failedCleanupProxy struct {
	err        error
	closeCalls int
}

func (*failedCleanupProxy) Host() string { return "" }

func (p *failedCleanupProxy) Close(context.Context) error {
	p.closeCalls++
	return p.err
}

func TestProviderCleanupReleasesFailedProxyAfterSandboxRemoval(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rmErr error
	}{
		{name: "removed"},
		{name: "already missing", rmErr: errors.New("sandbox 'owned' not found")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proxyErr := errors.New("proxy shutdown failed")
			proxy := &failedCleanupProxy{err: proxyErr}
			attempts := 0
			p := &Provider{started: true, ready: make(chan struct{}), name: "owned", proxy: proxy, run: func(_ context.Context, args ...string) ([]byte, error) {
				if strings.Join(args, " ") != "rm --force owned" {
					t.Fatalf("unexpected command: %v", args)
				}
				attempts++
				return nil, tc.rmErr
			}}
			close(p.ready)
			if err := p.Cleanup(context.Background()); !errors.Is(err, proxyErr) {
				t.Fatalf("expected original proxy shutdown error, got %v", err)
			}
			if p.name != "" || p.proxy != nil {
				t.Fatalf("retained removed sandbox resources: name=%q proxy=%v", p.name, p.proxy)
			}
			if proxy.closeCalls != 1 || attempts != 1 {
				t.Fatalf("expected one proxy close and removal, got close=%d rm=%d", proxy.closeCalls, attempts)
			}
			if err := p.Cleanup(context.Background()); err != nil {
				t.Fatalf("cleanup after confirmed removal failed: %v", err)
			}
			if proxy.closeCalls != 1 || attempts != 1 {
				t.Fatalf("cleanup repeated after confirmed removal: close=%d rm=%d", proxy.closeCalls, attempts)
			}
		})
	}
}

func TestProviderCleanupRetainsFailedProxyUntilSandboxRemoval(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rmErr error
	}{
		{name: "removed on retry"},
		{name: "already missing on retry", rmErr: errors.New("sandbox 'owned' not found")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proxyErr := errors.New("proxy shutdown failed")
			rmErr := errors.New("remove denied")
			proxy := &failedCleanupProxy{err: proxyErr}
			attempts := 0
			p := &Provider{started: true, ready: make(chan struct{}), name: "owned", proxy: proxy, run: func(_ context.Context, args ...string) ([]byte, error) {
				if strings.Join(args, " ") != "rm --force owned" {
					t.Fatalf("unexpected command: %v", args)
				}
				attempts++
				if attempts == 1 {
					return nil, rmErr
				}
				return nil, tc.rmErr
			}}
			close(p.ready)
			if err := p.Cleanup(context.Background()); !errors.Is(err, proxyErr) || !errors.Is(err, rmErr) {
				t.Fatalf("expected proxy shutdown and removal errors, got %v", err)
			}
			if p.name != "owned" || p.proxy != proxy {
				t.Fatalf("lost resources after removal failure: name=%q proxy=%v", p.name, p.proxy)
			}
			if proxy.closeCalls != 1 || attempts != 1 {
				t.Fatalf("expected one proxy close and removal, got close=%d rm=%d", proxy.closeCalls, attempts)
			}
			if err := p.Cleanup(context.Background()); !errors.Is(err, proxyErr) || errors.Is(err, rmErr) {
				t.Fatalf("expected only original proxy shutdown error after removal retry, got %v", err)
			}
			if p.name != "" || p.proxy != nil {
				t.Fatalf("retained removed sandbox resources: name=%q proxy=%v", p.name, p.proxy)
			}
			if proxy.closeCalls != 2 || attempts != 2 {
				t.Fatalf("expected retry of proxy close and removal, got close=%d rm=%d", proxy.closeCalls, attempts)
			}
			if err := p.Cleanup(context.Background()); err != nil {
				t.Fatalf("cleanup after confirmed removal failed: %v", err)
			}
			if proxy.closeCalls != 2 || attempts != 2 {
				t.Fatalf("cleanup repeated after confirmed removal: close=%d rm=%d", proxy.closeCalls, attempts)
			}
		})
	}
}

func TestProviderCleanupRetriesOwnedSandboxRemoval(t *testing.T) {
	failed := true
	attempts := 0
	p := &Provider{started: true, ready: make(chan struct{}), name: "owned", run: func(_ context.Context, args ...string) ([]byte, error) {
		if strings.Join(args, " ") != "rm --force owned" {
			t.Fatalf("unexpected command: %v", args)
		}
		attempts++
		if failed {
			return nil, errors.New("remove denied")
		}
		return nil, nil
	}}
	close(p.ready)
	if err := p.Cleanup(context.Background()); err == nil {
		t.Fatal("expected remove failure")
	}
	if p.name != "owned" {
		t.Fatalf("lost owned sandbox after removal failure: %q", p.name)
	}
	failed = false
	if err := p.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.name != "" {
		t.Fatalf("retained owned sandbox after confirmed removal: %q", p.name)
	}
	if err := p.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Fatalf("expected one failed removal and one successful retry, got %d attempts", attempts)
	}
}

// The fake CLI models a timed-out create after the runtime has allocated the
// sandbox. Cleanup must use a fresh context because the create context expired.
func TestAmbiguousCreateAttemptsCleanup(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "calls")
	binary := filepath.Join(dir, "fake-sbx")
	script := "#!/bin/sh\nprintf '%s\\n' \"$1\" >> '" + logPath + "'\nif [ \"$1\" = create ]; then exit 1; fi\n"
	if err := os.WriteFile(binary, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TESTCONTAINERS_SBX_BINARY", binary)
	t.Setenv("TESTCONTAINERS_SBX_WORKSPACE", dir)
	t.Setenv("TESTCONTAINERS_SBX_TIMEOUT", "2s")
	t.Setenv("TESTCONTAINERS_SBX_APP_NAME", "")
	p := &Provider{}
	_, err := p.Initialize(context.Background())
	if err == nil {
		t.Fatal("fake create unexpectedly succeeded")
	}
	data, readErr := os.ReadFile(logPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != "create\nrm\n" {
		t.Fatalf("ambiguous create must only create and clean up; commands: %q", data)
	}
}

func TestDetachedPromotionFailureCleansUpOwnedSandbox(t *testing.T) {
	for _, tc := range []struct {
		name       string
		mode       string
		timeout    string
		cleanupErr bool
	}{
		{name: "CLI failure", mode: "fail", timeout: "2s"},
		{name: "initialization timeout", mode: "timeout", timeout: "500ms"},
		{name: "cleanup failure joins errors", mode: "fail", timeout: "2s", cleanupErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			logPath := filepath.Join(dir, "calls")
			namePath := filepath.Join(dir, "owned-name")
			binary := filepath.Join(dir, "fake-sbx")
			script := `#!/bin/sh
printf '%s\n' "$1" >> "$TC_SBX_PROMOTION_CALLS"
case "$1" in
create)
    [ "$2" = --name ] && [ -n "$3" ] || exit 1
    printf '%s' "$3" > "$TC_SBX_PROMOTION_NAME"
    ;;
run)
    owned="$(cat "$TC_SBX_PROMOTION_NAME")"
    [ "$#" -eq 4 ] && [ "$2" = --detach ] && [ "$3" = --name ] && [ "$4" = "$owned" ] || exit 1
    if [ "$TC_SBX_PROMOTION_MODE" = timeout ]; then exec sleep 5; fi
    echo 'detached promotion denied' >&2
    exit 1
    ;;
rm)
    owned="$(cat "$TC_SBX_PROMOTION_NAME")"
    [ "$#" -eq 3 ] && [ "$2" = --force ] && [ "$3" = "$owned" ] || exit 1
    if [ "$TC_SBX_PROMOTION_CLEANUP_ERROR" = true ]; then
        echo 'remove denied' >&2
        exit 1
    fi
    ;;
*) echo 'unexpected discovery or other command' >&2; exit 1 ;;
esac
`
			if err := os.WriteFile(binary, []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("TESTCONTAINERS_SBX_BINARY", binary)
			t.Setenv("TESTCONTAINERS_SBX_WORKSPACE", dir)
			t.Setenv("TESTCONTAINERS_SBX_TIMEOUT", tc.timeout)
			t.Setenv("TESTCONTAINERS_SBX_APP_NAME", "")
			t.Setenv("TC_SBX_PROMOTION_CALLS", logPath)
			t.Setenv("TC_SBX_PROMOTION_NAME", namePath)
			t.Setenv("TC_SBX_PROMOTION_MODE", tc.mode)
			t.Setenv("TC_SBX_PROMOTION_CLEANUP_ERROR", strconv.FormatBool(tc.cleanupErr))
			factoryCalled := false
			p := &Provider{}
			p.proxyFactory = func(context.Context, sandboxRuntime) (proxyServer, error) {
				factoryCalled = true
				return nil, errors.New("unexpected proxy factory call")
			}
			start := time.Now()
			_, err := p.Initialize(context.Background())
			if err == nil || !strings.Contains(err.Error(), "sbx run") {
				t.Fatalf("lost detached promotion error: %v", err)
			}
			if tc.mode == "timeout" {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("expected initialization deadline, got %v", err)
				}
				if time.Since(start) > 3*time.Second {
					t.Fatal("detached promotion ignored initialization timeout")
				}
			} else if !strings.Contains(err.Error(), "detached promotion denied") {
				t.Fatalf("lost CLI diagnostic: %v", err)
			}
			if factoryCalled || p.proxy != nil {
				t.Fatal("proxy setup proceeded after failed detached promotion")
			}
			calls, readErr := os.ReadFile(logPath)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if string(calls) != "create\nrun\nrm\n" {
				t.Fatalf("expected promotion then fresh-context cleanup without discovery; commands: %q", calls)
			}
			owned, readErr := os.ReadFile(namePath)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if tc.cleanupErr {
				if !strings.Contains(err.Error(), "remove denied") || p.name != string(owned) {
					t.Fatalf("lost cleanup error or ownership: name=%q err=%v", p.name, err)
				}
			} else if p.name != "" {
				t.Fatalf("owned sandbox was not removed: %q (%v)", p.name, err)
			}
		})
	}
}

func TestBootstrapPublishesNumericDockerAPIPort(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "calls")
	publicationPath := filepath.Join(dir, "publication")
	binary := filepath.Join(dir, "fake-sbx")
	script := `#!/bin/sh
printf '%s\n' "$1" >> "$TC_SBX_BOOTSTRAP_CALLS"
case "$1" in
create)
    [ "$2" = --name ] && [ -n "$3" ] || exit 1
    printf '%s' "$3" > "$TC_SBX_BOOTSTRAP_NAME"
    previous=""
    for argument in "$@"; do
        if [ "$previous" = --kit ]; then printf '%s' "$argument" > "$TC_SBX_BOOTSTRAP_KIT"; fi
        previous="$argument"
    done
    shift
    while [ "$#" -gt 0 ]; do
        if [ "$1" = --publish ]; then
            shift
            spec="$1"
            break
        fi
        shift
    done
    case "$spec" in
        127.0.0.1:*:2375/tcp) ;;
        *) echo 'expected loopback Docker publication' >&2; exit 1 ;;
    esac
    port="${spec#127.0.0.1:}"
    port="${port%:2375/tcp}"
    case "$port" in
        ''|*[!0-9]*) echo 'expected explicit numeric host port' >&2; exit 1 ;;
    esac
    if [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
        echo 'host port out of range' >&2
        exit 1
    fi
    printf '%s' "$spec" > "$TC_SBX_BOOTSTRAP_PUBLICATION"
    ;;
run)
    owned="$(cat "$TC_SBX_BOOTSTRAP_NAME")"
    [ "$#" -eq 4 ] && [ "$2" = --detach ] && [ "$3" = --name ] && [ "$4" = "$owned" ] || exit 1
    ;;
ports)
    owned="$(cat "$TC_SBX_BOOTSTRAP_NAME")"
    [ "$#" -eq 4 ] && [ "$2" = "$owned" ] && [ "$3" = --format ] && [ "$4" = json ] || exit 1
    spec="$(cat "$TC_SBX_BOOTSTRAP_PUBLICATION")"
    port="${spec#127.0.0.1:}"
    port="${port%:2375/tcp}"
    kit_port=43220
    if [ "$port" -eq "$kit_port" ]; then kit_port=43221; fi
    printf '[{"host_ip":"127.0.0.1","host_port":%s,"sandbox_port":2375,"protocol":"tcp"},{"host_ip":"::1","host_port":%s,"sandbox_port":2375,"protocol":"tcp"}]' "$kit_port" "$kit_port"
    ;;
rm)
    owned="$(cat "$TC_SBX_BOOTSTRAP_NAME")"
    [ "$#" -eq 3 ] && [ "$2" = --force ] && [ "$3" = "$owned" ] || exit 1
    ;;
*) exit 1 ;;
esac
`
	if err := os.WriteFile(binary, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TESTCONTAINERS_SBX_BINARY", binary)
	t.Setenv("TESTCONTAINERS_SBX_WORKSPACE", dir)
	t.Setenv("TESTCONTAINERS_SBX_TIMEOUT", "2s")
	t.Setenv("TESTCONTAINERS_SBX_APP_NAME", "")
	t.Setenv("TC_SBX_BOOTSTRAP_CALLS", logPath)
	t.Setenv("TC_SBX_BOOTSTRAP_PUBLICATION", publicationPath)
	t.Setenv("TC_SBX_BOOTSTRAP_NAME", filepath.Join(dir, "owned-name"))
	kitPath := filepath.Join(dir, "kit-reference")
	t.Setenv("TC_SBX_BOOTSTRAP_KIT", kitPath)
	factoryCalled := false
	p := &Provider{}
	p.proxyFactory = func(context.Context, sandboxRuntime) (proxyServer, error) {
		factoryCalled = true
		return nil, errors.New("unexpected proxy factory call")
	}
	_, initializeErr := p.Initialize(context.Background())
	if initializeErr == nil {
		t.Fatal("bootstrap accepted unrelated kit publications")
	}
	if factoryCalled {
		t.Fatal("proxy factory called without the owned Docker publication")
	}
	if p.name != "" {
		t.Fatalf("discovery failure did not remove the owned sandbox: %q (%v)", p.name, initializeErr)
	}
	publication, err := os.ReadFile(publicationPath)
	if err != nil {
		t.Fatal(err)
	}
	address, ok := strings.CutSuffix(string(publication), ":2375/tcp")
	if !ok {
		t.Fatalf("unexpected Docker publication: %s", publication)
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil || host != "127.0.0.1" {
		t.Fatalf("Docker publication is not loopback: %s (%v)", publication, err)
	}
	value, err := strconv.Atoi(port)
	if err != nil || value < 1 || value > 65535 {
		t.Fatalf("Docker publication needs an explicit valid host port: %s", publication)
	}
	if !strings.Contains(initializeErr.Error(), "127.0.0.1:"+port) {
		t.Fatalf("missing requested endpoint in discovery error: %v", initializeErr)
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(calls) != "create\nrun\nports\nrm\n" {
		t.Fatalf("unexpected bootstrap/cleanup calls: %q", calls)
	}
	kitReference, err := os.ReadFile(kitPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(kitReference) != "sbx/docker-tcp:latest" {
		t.Fatalf("unexpected create --kit reference: %q", kitReference)
	}
}

func TestDiscoverDockerURLSelectsLoopbackEndpoint(t *testing.T) {
	run := func(context.Context, ...string) ([]byte, error) {
		return []byte(`[{"host_ip":"127.0.0.1","host_port":43219,"sandbox_port":2375,"protocol":"tcp"}]`), nil
	}
	u, err := discoverDockerURL(context.Background(), run, "owned", 43219)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != net.JoinHostPort("127.0.0.1", "43219") {
		t.Fatalf("unexpected endpoint: %s", u)
	}
}
