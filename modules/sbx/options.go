package sbx

import (
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"
)

// ErrProviderStarted is returned when configuration is attempted after provider
// initialization has begun, including after failed initialization or cleanup.
var ErrProviderStarted = errors.New("sbx provider initialization has started; configure before the first Testcontainers operation")

type options struct {
	kits      []string
	kitArgs   map[string]string
	binary    string
	workspace string
	timeout   time.Duration
	appName   *string
	bindRoots []string // nil inherits the environment; an empty slice clears it.
}

func (o options) clone() options {
	o.kits = slices.Clone(o.kits)
	o.kitArgs = maps.Clone(o.kitArgs)
	o.bindRoots = slices.Clone(o.bindRoots)
	return o
}

// Option configures a sandbox before provider initialization.
// These options apply to the sandbox, not to containers running inside it.
type Option func(*options) error

// Configure configures the registered default provider without starting or
// selecting it. Call it before any Testcontainers operation, typically in
// TestMain before m.Run or provider health checks. Select sbx separately through
// TESTCONTAINERS_PROVIDER or the provider property.
// Each call applies all its options atomically or leaves configuration unchanged.
// Explicit settings override TESTCONTAINERS_SBX_* environment variables, which
// are read when initialization begins.
func Configure(opts ...Option) error { return defaultProvider.Configure(opts...) }

// Configure configures this provider without starting it. Repeated calls compose
// until initialization begins; later calls return ErrProviderStarted.
// Each call applies all its options atomically or leaves configuration unchanged.
func (p *Provider) Configure(opts ...Option) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.started {
		return ErrProviderStarted
	}
	next := p.options.clone()
	for i, opt := range opts {
		if opt == nil {
			return fmt.Errorf("sbx option %d is nil", i)
		}
		if err := opt(&next); err != nil {
			return fmt.Errorf("configure sbx option %d: %w", i, err)
		}
	}
	p.options = next
	return nil
}

// WithBinary sets the sbx executable path or name, overriding
// TESTCONTAINERS_SBX_BINARY. The executable is resolved when initialization begins.
func WithBinary(binary string) Option {
	return func(o *options) error {
		if strings.TrimSpace(binary) == "" || strings.ContainsRune(binary, '\x00') {
			return fmt.Errorf("invalid sbx binary %q", binary)
		}
		o.binary = binary
		return nil
	}
}

// WithWorkspace sets the sandbox workspace, overriding TESTCONTAINERS_SBX_WORKSPACE.
// Relative paths are resolved against the process working directory at startup.
// The directory must exist when initialization begins.
func WithWorkspace(path string) Option {
	return func(o *options) error {
		if strings.TrimSpace(path) == "" || strings.ContainsRune(path, '\x00') {
			return fmt.Errorf("invalid sbx workspace %q", path)
		}
		o.workspace = path
		return nil
	}
}

// WithStartupTimeout sets the positive timeout for sandbox initialization,
// including kit installation, overriding TESTCONTAINERS_SBX_TIMEOUT.
func WithStartupTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		if timeout <= 0 {
			return fmt.Errorf("sbx startup timeout must be positive: %s", timeout)
		}
		o.timeout = timeout
		return nil
	}
}

// WithAppName sets the sbx CLI namespace for every command, overriding
// TESTCONTAINERS_SBX_APP_NAME. An empty name selects the CLI's default namespace.
// Names accept up to 20 ASCII letters, digits, hyphens, and underscores.
// Authenticate the selected namespace with sbx login before initialization.
func WithAppName(name string) Option {
	return func(o *options) error {
		if len(name) > 20 || strings.ContainsFunc(name, func(r rune) bool {
			return (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_'
		}) {
			return fmt.Errorf("invalid sbx app name %q: use up to 20 letters, digits, hyphens, or underscores", name)
		}
		o.appName = &name
		return nil
	}
}

// WithBindRoots replaces the additional allowed host bind roots, overriding
// TESTCONTAINERS_SBX_BIND_ROOTS and any earlier WithBindRoots option. Paths must be
// absolute directories that exist at startup. WithBindRoots() clears additional
// roots; the workspace is always allowed. Input paths are copied.
func WithBindRoots(paths ...string) Option {
	paths = append([]string{}, paths...)
	return func(o *options) error {
		for _, path := range paths {
			if !filepath.IsAbs(path) || strings.ContainsRune(path, '\x00') {
				return fmt.Errorf("invalid sbx bind root %q: path must be absolute and contain no NUL bytes", path)
			}
		}
		o.bindRoots = slices.Clone(paths)
		return nil
	}
}

// WithKits appends mixin references to the sandbox's creation request, in order.
// The required Docker TCP kit is always included automatically. sbx validates
// source references, duplicate kits, dependencies, and compatibility.
// References may be OCI images, local directories, or Git sources. Relative paths
// are resolved by the CLI against the process working directory at startup,
// independently of WithWorkspace or TESTCONTAINERS_SBX_WORKSPACE.
func WithKits(refs ...string) Option {
	refs = slices.Clone(refs)
	return func(o *options) error {
		for _, ref := range refs {
			if strings.TrimSpace(ref) == "" || strings.ContainsRune(ref, '\x00') {
				return fmt.Errorf("invalid sbx kit reference %q", ref)
			}
		}
		o.kits = append(o.kits, refs...)
		return nil
	}
}

// WithKitArgs merges kit arguments. Later values override earlier values for the
// same key. Use "name" for a shared argument or "kit-handle.name" to target one
// kit. sbx validates declared arguments and their values. Empty values are valid.
// Arguments are plain text recorded by sbx; use sbx credentials for secrets.
func WithKitArgs(args map[string]string) Option {
	args = maps.Clone(args)
	return func(o *options) error {
		for _, key := range slices.Sorted(maps.Keys(args)) {
			if key == "" || strings.ContainsAny(key, "=\x00") || strings.ContainsFunc(key, unicode.IsSpace) {
				return fmt.Errorf("invalid sbx kit argument key %q", key)
			}
			if strings.ContainsRune(args[key], '\x00') {
				return fmt.Errorf("sbx kit argument %q contains a NUL byte", key)
			}
		}
		if o.kitArgs == nil {
			o.kitArgs = make(map[string]string, len(args))
		}
		maps.Copy(o.kitArgs, args)
		return nil
	}
}
