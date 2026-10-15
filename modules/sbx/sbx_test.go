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

func TestProviderCleanupRetainsKitUntilRemovalConfirmed(t *testing.T) {
	dir := t.TempDir()
	kit := filepath.Join(dir, "kit")
	if err := os.Mkdir(kit, 0o700); err != nil {
		t.Fatal(err)
	}
	failed := true
	p := &Provider{started: true, ready: make(chan struct{}), name: "owned", kitDir: kit, run: func(_ context.Context, args ...string) ([]byte, error) {
		if len(args) < 1 || args[0] != "rm" {
			t.Fatalf("unexpected command: %v", args)
		}
		if failed {
			return nil, errors.New("remove denied")
		}
		return nil, nil
	}}
	close(p.ready)
	if err := p.Cleanup(context.Background()); err == nil {
		t.Fatal("expected remove failure")
	}
	if _, err := os.Stat(kit); err != nil {
		t.Fatalf("kit removed before sandbox: %v", err)
	}
	failed = false
	if err := p.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(kit); !os.IsNotExist(err) {
		t.Fatalf("kit survived confirmed removal: %v", err)
	}
	if err := p.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
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
	if !strings.Contains(string(data), "rm\n") {
		t.Fatalf("ambiguous create was not cleaned up; commands: %q", data)
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
ports)
    spec="$(cat "$TC_SBX_BOOTSTRAP_PUBLICATION")"
    port="${spec#127.0.0.1:}"
    port="${port%:2375/tcp}"
    kit_port=43220
    if [ "$port" -eq "$kit_port" ]; then kit_port=43221; fi
    printf '[{"host_ip":"127.0.0.1","host_port":%s,"sandbox_port":2375,"protocol":"tcp"},{"host_ip":"::1","host_port":%s,"sandbox_port":2375,"protocol":"tcp"}]' "$kit_port" "$kit_port"
    ;;
rm) ;;
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
	kitPath := filepath.Join(dir, "kit-path")
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
	if string(calls) != "create\nports\nrm\n" {
		t.Fatalf("unexpected bootstrap/cleanup calls: %q", calls)
	}
	kitDir, err := os.ReadFile(kitPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(kitDir) == 0 {
		t.Fatal("create did not receive the mandatory kit")
	}
	if _, err := os.Stat(string(kitDir)); !os.IsNotExist(err) {
		t.Fatalf("kit survived failed endpoint discovery: %v", err)
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
