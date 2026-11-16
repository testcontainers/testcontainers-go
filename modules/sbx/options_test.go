package sbx

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConfigureComposesAndCopiesOptions(t *testing.T) {
	refs := []string{"./test-tools"}
	args := map[string]string{"test-tools.mode": "original"}
	kits, kitArgs := WithKits(refs...), WithKitArgs(args)
	refs[0], args["test-tools.mode"] = "changed", "changed"
	var p, other Provider
	for _, provider := range []*Provider{&p, &other} {
		if err := provider.Configure(kits, kitArgs); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.Configure(WithKits("sbx/another:1.0"), WithKitArgs(map[string]string{
		"test-tools.mode": "integration", "empty": "",
	})); err != nil {
		t.Fatal(err)
	}
	want := options{
		kits:    []string{"./test-tools", "sbx/another:1.0"},
		kitArgs: map[string]string{"test-tools.mode": "integration", "empty": ""},
	}
	if !reflect.DeepEqual(p.options, want) {
		t.Fatalf("configured options: got %#v, want %#v", p.options, want)
	}
	if other.options.kits[0] != "./test-tools" || other.options.kitArgs["test-tools.mode"] != "original" {
		t.Fatalf("reusing options shared mutable state: %#v", other.options)
	}
	if p.started || other.started {
		t.Fatal("configuration started a provider")
	}
}

func TestConfigureRollsBackInvalidOptions(t *testing.T) {
	for _, tc := range []struct {
		name string
		opt  Option
	}{
		{name: "nil", opt: nil},
		{name: "empty reference", opt: WithKits("")},
		{name: "blank reference", opt: WithKits(" \t\n")},
		{name: "NUL reference", opt: WithKits("kit\x00")},
		{name: "empty key", opt: WithKitArgs(map[string]string{"": "value"})},
		{name: "equals key", opt: WithKitArgs(map[string]string{"mode=value": "value"})},
		{name: "whitespace key", opt: WithKitArgs(map[string]string{"mode\n": "value"})},
		{name: "NUL key", opt: WithKitArgs(map[string]string{"mode\x00": "value"})},
		{name: "NUL value", opt: WithKitArgs(map[string]string{"mode": "value\x00"})},
		{name: "empty-binary", opt: WithBinary("")},
		{name: "NUL-binary", opt: WithBinary("sbx\x00")},
		{name: "empty-workspace", opt: WithWorkspace("")},
		{name: "NUL-workspace", opt: WithWorkspace("/tmp\x00")},
		{name: "zero-timeout", opt: WithStartupTimeout(0)},
		{name: "negative-timeout", opt: WithStartupTimeout(-time.Second)},
		{name: "long-app-name", opt: WithAppName(strings.Repeat("a", 21))},
		{name: "invalid-app-name", opt: WithAppName("test suite")},
		{name: "relative-bind-root", opt: WithBindRoots("relative")},
		{name: "empty-bind-root", opt: WithBindRoots("")},
		{name: "NUL-bind-root", opt: WithBindRoots("/tmp\x00")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var p Provider
			if err := p.Configure(WithKits("./original"), WithKitArgs(map[string]string{"mode": "original"})); err != nil {
				t.Fatal(err)
			}
			before := p.options.clone()
			if err := p.Configure(WithKits("./new"), WithKitArgs(map[string]string{"mode": "new"}),
				WithWorkspace("./new"), WithAppName("new"), WithBindRoots(), tc.opt); err == nil {
				t.Fatal("invalid option accepted")
			}
			if !reflect.DeepEqual(p.options, before) {
				t.Fatalf("failed configuration mutated options: %#v", p.options)
			}
			if err := p.Configure(WithKits("./valid")); err != nil {
				t.Fatalf("invalid option prevented subsequent configuration: %v", err)
			}
		})
	}
}

func TestConfigureRacesWithInitialization(t *testing.T) {
	// Fail before invoking the CLI; initialization still freezes configuration.
	t.Setenv("TESTCONTAINERS_SBX_WORKSPACE", t.TempDir())
	t.Setenv("TESTCONTAINERS_SBX_BIND_ROOTS", "")
	t.Setenv("TESTCONTAINERS_SBX_TIMEOUT", "invalid")
	for range 100 {
		var p Provider
		start := make(chan struct{})
		configured := make(chan error, 1)
		initialized := make(chan error, 1)
		go func() {
			<-start
			configured <- p.Configure(WithKits("./test-tools"))
		}()
		go func() {
			<-start
			_, err := p.Initialize(t.Context())
			initialized <- err
		}()
		close(start)
		configureErr, initializeErr := <-configured, <-initialized
		if initializeErr == nil {
			t.Fatal("invalid timeout accepted")
		}
		if configureErr == nil {
			if !slices.Equal(p.options.kits, []string{"./test-tools"}) {
				t.Fatal("successful configuration lost")
			}
		} else if !errors.Is(configureErr, ErrProviderStarted) || len(p.options.kits) != 0 {
			t.Fatalf("configuration did not reject atomically: options=%#v error=%v", p.options, configureErr)
		}
		if err := p.Cleanup(t.Context()); err != nil {
			t.Fatal(err)
		}
		if err := p.Configure(WithKits("./late")); !errors.Is(err, ErrProviderStarted) {
			t.Fatalf("configuration after failure and cleanup: %v", err)
		}
	}
}

func TestCanceledInitializeDoesNotFreezeConfiguration(t *testing.T) {
	var p Provider
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := p.Initialize(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled initialization: %v", err)
	}
	if err := p.Configure(WithKits("./test-tools")); err != nil {
		t.Fatalf("canceled caller froze configuration: %v", err)
	}
}

func TestKitCreateArgumentsAndFailureCleanup(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "fake-sbx")
	script := `#!/bin/sh
[ "$1" = --app-name ] && [ "$2" = kit-tests ] || exit 1
shift 2
printf '%s\n' "$1" >> "$TC_SBX_KIT_CALLS"
case "$1" in
create)
    printf '%s\000' "$@" > "$TC_SBX_KIT_ARGS"
    printf '%s' "$3" > "$TC_SBX_KIT_NAME"
    echo 'kit composition failed' >&2
    exit 1
    ;;
rm)
    [ "$#" -eq 3 ] && [ "$2" = --force ] && [ "$3" = "$(cat "$TC_SBX_KIT_NAME")" ] || exit 1
    ;;
*) echo 'unexpected command after kit failure' >&2; exit 1 ;;
esac
`
	if err := os.WriteFile(binary, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TESTCONTAINERS_SBX_BINARY", "does-not-exist")
	t.Setenv("TESTCONTAINERS_SBX_WORKSPACE", filepath.Join(dir, "missing"))
	t.Setenv("TESTCONTAINERS_SBX_BIND_ROOTS", "relative")
	t.Setenv("TESTCONTAINERS_SBX_APP_NAME", "wrong-namespace")
	t.Setenv("TESTCONTAINERS_SBX_TIMEOUT", "invalid")
	argsPath, callsPath := filepath.Join(dir, "args"), filepath.Join(dir, "calls")
	t.Setenv("TC_SBX_KIT_ARGS", argsPath)
	t.Setenv("TC_SBX_KIT_CALLS", callsPath)
	t.Setenv("TC_SBX_KIT_NAME", filepath.Join(dir, "name"))
	var p Provider
	gitRef := "git+https://github.com/example/kits.git#ref=abc&dir=tools"
	value := "spaces = values\n$(not-a-command) `nor-this`"
	if err := p.Configure(
		WithBinary(binary), WithWorkspace(dir), WithBindRoots(),
		WithAppName("kit-tests"), WithStartupTimeout(5*time.Second),
		WithKits("./kit with spaces", gitRef, "docker.io/example/tools:1.0"),
		WithKitArgs(map[string]string{"tools.mode": value, "empty": ""}),
	); err != nil {
		t.Fatal(err)
	}
	// Concurrent clients must share one create attempt and its error.
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			_, err := p.Initialize(t.Context())
			if err == nil || !strings.Contains(err.Error(), "kit composition failed") {
				t.Errorf("lost kit error: %v", err)
			}
		})
	}
	wg.Wait()
	data, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Split(strings.TrimSuffix(string(data), "\x00"), "\x00")
	if len(args) != 19 {
		t.Fatalf("unexpected create arguments: %q", args)
	}
	want := []string{
		"create", "--name", args[2], "--kit", dockerTCPKit,
		"--kit", "./kit with spaces", "--kit", gitRef, "--kit", "docker.io/example/tools:1.0",
		"--kit-arg", "empty=", "--kit-arg", "tools.mode=" + value,
		"--publish", args[16], baseImage, dir,
	}
	if !slices.Equal(args, want) {
		t.Fatalf("create arguments: got %q, want %q", args, want)
	}
	data, err = os.ReadFile(callsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "create\nrm\n" || p.name != "" {
		t.Fatalf("kit failure cleanup: calls=%q owned=%q", data, p.name)
	}
	if err := p.Configure(WithKits("./late")); !errors.Is(err, ErrProviderStarted) {
		t.Fatalf("late configuration: %v", err)
	}
}

func TestConfigureCompletesBeforeInitialization(t *testing.T) {
	var p Provider
	// Hold an option application to prove Initialize waits for the full batch.
	entered, release := make(chan struct{}), make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	defer unblock()
	configured := make(chan error, 1)
	go func() {
		configured <- p.Configure(func(o *options) error {
			close(entered)
			<-release
			o.kits = append(o.kits, "./test-tools")
			return nil
		})
	}()
	<-entered
	// The canceled caller test covers the non-starting path; here force an actual
	// initialization attempt that fails locally, without needing sbx installed.
	t.Setenv("TESTCONTAINERS_SBX_WORKSPACE", t.TempDir())
	t.Setenv("TESTCONTAINERS_SBX_BIND_ROOTS", "")
	t.Setenv("TESTCONTAINERS_SBX_TIMEOUT", "invalid")
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	initialized := make(chan error, 1)
	go func() {
		_, err := p.Initialize(ctx)
		initialized <- err
	}()
	unblock()
	if err := <-configured; err != nil {
		t.Fatal(err)
	}
	if err := <-initialized; err == nil {
		t.Fatal("invalid timeout accepted")
	}
	if !slices.Equal(p.options.kits, []string{"./test-tools"}) {
		t.Fatal("initialization lost configuration")
	}
}
