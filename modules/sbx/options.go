package sbx

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"
)

// ErrProviderStarted is returned when configuration is attempted after provider
// initialization has begun, including after failed initialization or cleanup.
var ErrProviderStarted = errors.New("sbx provider initialization has started; configure before the first Testcontainers operation")

type options struct {
	kits    []string
	kitArgs map[string]string
}

func (o options) clone() options {
	return options{kits: slices.Clone(o.kits), kitArgs: maps.Clone(o.kitArgs)}
}

// Option configures a sandbox before provider initialization.
// These options apply to the sandbox, not to containers running inside it.
type Option func(*options) error

// Configure configures the registered default provider without starting or
// selecting it. Call it before any Testcontainers operation, typically in
// TestMain before m.Run or provider health checks. Select sbx separately through
// TESTCONTAINERS_PROVIDER or the provider property.
// Each call applies all its options atomically or leaves configuration unchanged.
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

// WithKits appends mixin references to the sandbox's creation request, in order.
// The required Docker TCP kit is always included automatically. sbx validates
// source references, duplicate kits, dependencies, and compatibility.
// References may be OCI images, local directories, or Git sources. Relative paths
// are resolved by the CLI against the process working directory at startup,
// independently of TESTCONTAINERS_SBX_WORKSPACE.
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
