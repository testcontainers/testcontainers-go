package sbx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	dockerSocketPath = "/var/run/docker.sock"
	mountTimeout     = 30 * time.Second
	maxBindSymlinks  = 255
)

type requestedBind struct {
	source   string
	readOnly bool
}

type mountedPath struct {
	hostPath string
	info     os.FileInfo
	readOnly bool
}

type mountRollback struct {
	source   string
	previous *mountedPath
}

type authorizedBindRoot struct {
	configuredBindRoot
	root *os.Root
	info os.FileInfo
}

type resolvedBind struct {
	hostPath         string
	info             os.FileInfo
	root             *authorizedBindRoot
	visibleWorkspace bool
}

// bindMountManager serializes runtime mounts. Besides avoiding duplicate CLI
// calls, the file identity in the cache makes a deleted and recreated temp
// path mount again rather than continuing to refer to the old inode.
type bindMountManager struct {
	runtime sandboxRuntime

	mu      sync.Mutex
	mounted map[string]mountedPath
	roots   []authorizedBindRoot
	closed  bool
	pending []mountRollback
}

func newBindMountManager(runtime sandboxRuntime) (*bindMountManager, error) {
	m := &bindMountManager{runtime: runtime, mounted: make(map[string]mountedPath)}
	for _, configured := range runtime.bindRoots {
		root, err := os.OpenRoot(configured.canonical)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("open sbx bind root %q: %w", configured.path, err), m.closeRoots())
		}
		info, err := root.Stat(".")
		if err != nil {
			_ = root.Close()
			return nil, errors.Join(fmt.Errorf("inspect sbx bind root %q: %w", configured.path, err), m.closeRoots())
		}
		m.roots = append(m.roots, authorizedBindRoot{configuredBindRoot: configured, root: root, info: info})
	}
	if len(m.roots) == 0 {
		return nil, errors.New("initialize sbx bind mounts: no authorized roots")
	}
	return m, nil
}

func (m *bindMountManager) prepare(ctx context.Context, host map[string]json.RawMessage) error {
	requests, err := requestedBindMounts(host)
	if err != nil {
		return err
	}
	sources := make([]string, 0, len(requests))
	for source := range requests {
		sources = append(sources, source)
	}
	slices.Sort(sources)

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("prepare sbx bind mounts: manager is closed")
	}
	if err := m.reconcileRollbacks(); err != nil {
		return err
	}
	var changes []mountRollback
	for _, source := range sources {
		request := requests[source]
		change, err := m.prepareOne(ctx, request)
		if change != nil {
			changes = append(changes, *change)
		}
		if err != nil {
			m.pending = changes
			return errors.Join(err, m.reconcileRollbacks())
		}
	}
	return nil
}

// Failed batches retain their undo records until cleanup succeeds. Retry them
// before preparing another batch so a leftover mount cannot be reused as if
// its create had succeeded. The caller holds m.mu throughout both operations.
func (m *bindMountManager) reconcileRollbacks() error {
	if len(m.pending) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), mountTimeout)
	defer cancel()
	var pending []mountRollback
	var errs []error
	for i := len(m.pending) - 1; i >= 0; i-- {
		change := m.pending[i]
		if err := m.rollbackMount(ctx, change); err != nil {
			pending = append(pending, change)
			errs = append(errs, fmt.Errorf("rollback sbx bind source %q: %w", change.source, err))
		}
	}
	slices.Reverse(pending)
	m.pending = pending
	return errors.Join(errs...)
}

func (m *bindMountManager) rollbackMount(ctx context.Context, change mountRollback) error {
	if current, ok := m.mounted[change.source]; ok {
		if _, err := m.runtime.run(ctx, "umount", m.runtime.name, current.hostPath+":"+change.source); err != nil {
			return err
		}
		delete(m.mounted, change.source)
	}
	if change.previous == nil {
		return nil
	}
	previous := *change.previous
	resolved, err := m.authorize(requestedBind{source: previous.hostPath, readOnly: previous.readOnly})
	if err != nil {
		return err
	}
	if !os.SameFile(previous.info, resolved.info) {
		return errors.New("previous mount source changed before restoration")
	}
	return m.mount(ctx, change.source, previous)
}

func (m *bindMountManager) validate(host map[string]json.RawMessage) error {
	requests, err := requestedBindMounts(host)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("validate sbx bind mounts: manager is closed")
	}
	for _, request := range requests {
		if _, err := m.authorize(request); err != nil {
			return err
		}
	}
	return nil
}

func (m *bindMountManager) authorize(request requestedBind) (resolvedBind, error) {
	source := filepath.Clean(request.source)
	if source == dockerSocketPath {
		// Ryuk and other Docker-aware containers must receive the sandbox's
		// Docker socket. Mounting the host socket here would escape the runtime.
		return resolvedBind{}, nil
	}
	if !filepath.IsAbs(source) {
		return resolvedBind{}, fmt.Errorf("sbx bind source %q must be an absolute host path", request.source)
	}

	resolved, err := m.resolve(source)
	if err != nil {
		if os.IsNotExist(err) {
			return resolvedBind{}, fmt.Errorf("sbx bind source %q does not exist on the host", source)
		}
		return resolvedBind{}, fmt.Errorf("authorize sbx bind source %q: %w", source, err)
	}
	info := resolved.info
	if info.Mode()&os.ModeSocket != 0 {
		return resolvedBind{}, fmt.Errorf("sbx cannot expose host socket %q; only %s is supported and resolves to the sandbox Docker socket", source, dockerSocketPath)
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return resolvedBind{}, fmt.Errorf("sbx bind source %q has unsupported file type %s", source, info.Mode().Type())
	}
	if err := verifyBindRoot(resolved.root); err != nil {
		return resolvedBind{}, fmt.Errorf("verify sbx bind root for source %q: %w", source, err)
	}
	return resolved, nil
}

func (m *bindMountManager) prepareOne(ctx context.Context, request requestedBind) (*mountRollback, error) {
	resolved, err := m.authorize(request)
	if err != nil {
		return nil, err
	}
	source := filepath.Clean(request.source)
	if source == dockerSocketPath {
		return nil, nil
	}
	info := resolved.info

	var change *mountRollback
	if mounted, ok := m.mounted[source]; ok {
		sameFile := mounted.hostPath == resolved.hostPath && os.SameFile(mounted.info, info)
		permissionsSuffice := !mounted.readOnly || request.readOnly
		if !resolved.visibleWorkspace && sameFile && permissionsSuffice {
			return nil, nil
		}
		unmountCtx, cancel := boundedContext(ctx, mountTimeout)
		_, unmountErr := m.runtime.run(unmountCtx, "umount", m.runtime.name, mounted.hostPath+":"+source)
		cancel()
		if unmountErr != nil {
			return nil, fmt.Errorf("refresh sbx bind source %q: %w", source, unmountErr)
		}
		delete(m.mounted, source)
		change = &mountRollback{source: source}
		// Permission upgrades can be undone. A stale inode or retargeted
		// alias cannot reliably be restored through its old pathname.
		if sameFile {
			change.previous = &mounted
		}
	}
	if resolved.visibleWorkspace {
		return change, nil
	}

	// The confined lookup cannot pass its descriptor through the sbx CLI. Use
	// the canonical path it verified, while keeping Docker's guest path explicit.
	// A later pathname replacement can still race the CLI's mount operation.
	if err := m.mount(ctx, source, mountedPath{hostPath: resolved.hostPath, info: info, readOnly: request.readOnly}); err != nil {
		return change, err
	}
	if change == nil {
		change = &mountRollback{source: source}
	}
	return change, nil
}

func (m *bindMountManager) mount(ctx context.Context, source string, mounted mountedPath) error {
	spec := mounted.hostPath + ":" + source
	if mounted.readOnly {
		spec += ":ro"
	}
	mountCtx, cancel := boundedContext(ctx, mountTimeout)
	_, err := m.runtime.run(mountCtx, "mount", m.runtime.name, spec)
	cancel()
	if err != nil {
		return fmt.Errorf("make host bind source %q visible in sandbox %q: %w", source, m.runtime.name, err)
	}
	m.mounted[source] = mounted
	return nil
}

func (m *bindMountManager) resolve(source string) (resolvedBind, error) {
	type candidate struct {
		root             *authorizedBindRoot
		relative         string
		workspaceVisible bool
	}
	var candidates []candidate
	for i := range m.roots {
		root := &m.roots[i]
		aliases := []string{root.path}
		if root.canonical != root.path {
			aliases = append(aliases, root.canonical)
		}
		for _, alias := range aliases {
			relative, ok := relativeWithin(alias, source)
			if !ok {
				continue
			}
			candidates = append(candidates, candidate{
				root:             root,
				relative:         relative,
				workspaceVisible: root.workspace && alias == root.path,
			})
		}
	}
	if len(candidates) == 0 {
		return resolvedBind{}, errors.New("path is outside the configured workspace and TESTCONTAINERS_SBX_BIND_ROOTS")
	}
	slices.SortStableFunc(candidates, func(a, b candidate) int {
		return len(b.root.canonical) - len(a.root.canonical)
	})
	var first resolvedBind
	var firstErr error
	for _, candidate := range candidates {
		relative, info, err := resolveInRoot(candidate.root.root, candidate.relative)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		resolved := resolvedBind{
			hostPath:         filepath.Join(candidate.root.canonical, relative),
			info:             info,
			root:             candidate.root,
			visibleWorkspace: candidate.workspaceVisible,
		}
		if resolved.visibleWorkspace {
			return resolved, nil
		}
		if first.info == nil {
			first = resolved
		}
	}
	if first.info != nil {
		return first, nil
	}
	return resolvedBind{}, firstErr
}

func verifyBindRoot(root *authorizedBindRoot) error {
	current, err := os.Stat(root.canonical)
	if err != nil {
		return err
	}
	if !os.SameFile(root.info, current) {
		return errors.New("configured root path changed after initialization")
	}
	return nil
}

func relativeWithin(root, path string) (string, bool) {
	if root == "" {
		return "", false
	}
	relative, err := filepath.Rel(root, path)
	return relative, err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func pathWithin(root, path string) bool {
	_, ok := relativeWithin(root, path)
	return ok
}

func resolveInRoot(root *os.Root, name string) (string, os.FileInfo, error) {
	if filepath.IsAbs(name) {
		return "", nil, errors.New("root-relative bind path is absolute")
	}
	pending := pathComponents(filepath.Clean(name))
	resolved := make([]string, 0, len(pending))
	links := 0
	for len(pending) > 0 {
		component := pending[0]
		pending = pending[1:]
		switch component {
		case ".", "":
			continue
		case "..":
			if len(resolved) == 0 {
				return "", nil, errors.New("relative symbolic link escapes its authorized root")
			}
			resolved = resolved[:len(resolved)-1]
			continue
		}

		candidate := filepath.Join(append(slices.Clone(resolved), component)...)
		info, err := root.Lstat(candidate)
		if err != nil {
			return "", nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			links++
			if links > maxBindSymlinks {
				return "", nil, errors.New("too many symbolic links in bind source")
			}
			target, err := root.Readlink(candidate)
			if err != nil {
				return "", nil, err
			}
			if filepath.IsAbs(target) {
				return "", nil, errors.New("absolute symbolic links are not allowed in bind sources")
			}
			pending = append(pathComponents(target), pending...)
			continue
		}
		if len(pending) > 0 && !info.IsDir() {
			return "", nil, fmt.Errorf("bind source component %q is not a directory", candidate)
		}
		resolved = append(resolved, component)
	}

	relative := "."
	if len(resolved) > 0 {
		relative = filepath.Join(resolved...)
	}
	info, err := root.Stat(relative)
	if err != nil {
		return "", nil, err
	}
	return relative, info, nil
}

func pathComponents(path string) []string {
	if path == "." || path == "" {
		return nil
	}
	return strings.Split(path, string(filepath.Separator))
}

func (m *bindMountManager) close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	m.closed = true
	return errors.Join(m.reconcileRollbacks(), m.closeRoots())
}

func (m *bindMountManager) closeRoots() error {
	var errs []error
	for i := range m.roots {
		if err := m.roots[i].root.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close sbx bind root %q: %w", m.roots[i].path, err))
		}
	}
	m.roots = nil
	return errors.Join(errs...)
}

func canonicalizeDockerMounts(host *dockerJSONObject) error {
	raw, ok := host.get("Mounts")
	if !ok || string(raw) == "null" {
		return nil
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		return fmt.Errorf("decode Docker HostConfig.Mounts: %w", err)
	}
	for i, entry := range entries {
		if string(entry) == "null" {
			continue
		}
		mount, err := decodeDockerJSONObject(entry, fmt.Sprintf("Docker HostConfig.Mounts[%d]", i), "Type", "Source", "ReadOnly")
		if err != nil {
			return fmt.Errorf("decode Docker HostConfig.Mounts: %w", err)
		}
		mount.canonicalize()
		canonical, err := json.Marshal(mount)
		if err != nil {
			return fmt.Errorf("encode Docker HostConfig.Mounts[%d]: %w", i, err)
		}
		entries[i] = canonical
	}
	canonical, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("encode Docker HostConfig.Mounts: %w", err)
	}
	host.set("Mounts", canonical)
	return nil
}

func requestedBindMounts(host map[string]json.RawMessage) (map[string]requestedBind, error) {
	requests := make(map[string]requestedBind)
	if raw := host["Binds"]; len(raw) > 0 && string(raw) != "null" {
		var binds []string
		if err := json.Unmarshal(raw, &binds); err != nil {
			return nil, fmt.Errorf("decode Docker HostConfig.Binds: %w", err)
		}
		for _, bind := range binds {
			source, readOnly, filesystem, err := parseDockerBind(bind)
			if err != nil {
				return nil, err
			}
			if filesystem {
				if err := validateBindSourceSyntax(source); err != nil {
					return nil, err
				}
				mergeRequestedBind(requests, source, readOnly)
			}
		}
	}
	if raw := host["Mounts"]; len(raw) > 0 && string(raw) != "null" {
		var mounts []struct {
			Type     string `json:"Type"`
			Source   string `json:"Source"`
			ReadOnly bool   `json:"ReadOnly"`
		}
		if err := json.Unmarshal(raw, &mounts); err != nil {
			return nil, fmt.Errorf("decode Docker HostConfig.Mounts: %w", err)
		}
		for _, mount := range mounts {
			if !strings.EqualFold(mount.Type, "bind") {
				continue
			}
			if mount.Source == "" {
				return nil, errors.New("docker bind mount has an empty source")
			}
			if err := validateBindSourceSyntax(mount.Source); err != nil {
				return nil, err
			}
			mergeRequestedBind(requests, mount.Source, mount.ReadOnly)
		}
	}
	return requests, nil
}

func validateBindSourceSyntax(source string) error {
	for _, component := range strings.Split(source, string(filepath.Separator)) {
		if component == ".." {
			return fmt.Errorf("sbx bind source %q must not contain '..' path components", source)
		}
	}
	return nil
}

func parseDockerBind(spec string) (source string, readOnly, filesystem bool, err error) {
	// A drive-letter source with its own destination is a Windows host bind,
	// which this adapter cannot expose through its POSIX guest paths. Keep
	// C:/data available as the valid named volume C mounted at /data.
	if len(spec) >= 3 && ((spec[0] >= 'A' && spec[0] <= 'Z') || (spec[0] >= 'a' && spec[0] <= 'z')) &&
		spec[1] == ':' && (spec[2] == '\\' || (spec[2] == '/' && strings.Contains(spec[3:], ":"))) {
		return "", false, false, fmt.Errorf("sbx supports only POSIX absolute host bind paths; Windows drive-letter bind %q is unsupported", spec)
	}
	parts := strings.SplitN(spec, ":", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", false, false, fmt.Errorf("invalid Docker bind %q: expected source:destination[:options]", spec)
	}
	// The Binds API also represents named volumes. Only absolute sources are
	// host filesystem paths that need sandbox visibility.
	if !filepath.IsAbs(parts[0]) {
		return parts[0], false, false, nil
	}
	if len(parts) == 3 {
		for _, option := range strings.Split(parts[2], ",") {
			if option == "ro" {
				readOnly = true
			}
		}
	}
	return parts[0], readOnly, true, nil
}

func mergeRequestedBind(requests map[string]requestedBind, source string, readOnly bool) {
	source = filepath.Clean(source)
	current, ok := requests[source]
	if !ok {
		requests[source] = requestedBind{source: source, readOnly: readOnly}
		return
	}
	// A read-write outer mount satisfies both request forms; Docker still
	// enforces the individual container's ReadOnly flag on its inner bind.
	current.readOnly = current.readOnly && readOnly
	requests[source] = current
}
