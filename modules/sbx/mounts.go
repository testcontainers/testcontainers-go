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

// bindMountManager serializes runtime mounts. Besides avoiding duplicate CLI
// calls, the file identity in the cache makes a deleted and recreated temp
// path mount again rather than continuing to refer to the old inode.
type bindMountManager struct {
	runtime sandboxRuntime

	mu      sync.Mutex
	mounted map[string]mountedPath
}

func newBindMountManager(runtime sandboxRuntime) *bindMountManager {
	return &bindMountManager{runtime: runtime, mounted: make(map[string]mountedPath)}
}

func (m *bindMountManager) prepare(ctx context.Context, host map[string]json.RawMessage) error {
	requests, err := requestedBindMounts(host)
	if err != nil {
		return err
	}
	if len(requests) == 0 {
		return nil
	}

	workspace := ""
	if m.runtime.workspace != "" {
		workspace, err = filepath.Abs(m.runtime.workspace)
		if err != nil {
			return fmt.Errorf("resolve configured sbx workspace %q: %w", m.runtime.workspace, err)
		}
	}

	sources := make([]string, 0, len(requests))
	for source := range requests {
		sources = append(sources, source)
	}
	slices.Sort(sources)

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, source := range sources {
		request := requests[source]
		if err := m.prepareOne(ctx, workspace, request); err != nil {
			return err
		}
	}
	return nil
}

func (m *bindMountManager) prepareOne(ctx context.Context, workspace string, request requestedBind) error {
	source := filepath.Clean(request.source)
	if source == dockerSocketPath {
		// Ryuk and other Docker-aware containers must receive the sandbox's
		// Docker socket. Mounting the host socket here would escape the runtime.
		return nil
	}
	if !filepath.IsAbs(source) {
		return fmt.Errorf("sbx bind source %q must be an absolute host path", request.source)
	}

	info, err := os.Stat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("sbx bind source %q does not exist on the host", source)
		}
		return fmt.Errorf("stat sbx bind source %q: %w", source, err)
	}
	if info.Mode()&os.ModeSocket != 0 {
		return fmt.Errorf("sbx cannot expose host socket %q; only %s is supported and resolves to the sandbox Docker socket", source, dockerSocketPath)
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return fmt.Errorf("sbx bind source %q has unsupported file type %s", source, info.Mode().Type())
	}

	resolved, resolveErr := filepath.EvalSymlinks(source)
	if resolveErr != nil {
		return fmt.Errorf("resolve sbx bind source %q: %w", source, resolveErr)
	}
	visibleInWorkspace := workspacePathVisible(workspace, source)

	if mounted, ok := m.mounted[source]; ok {
		sameFile := mounted.hostPath == resolved && os.SameFile(mounted.info, info)
		permissionsSuffice := !mounted.readOnly || request.readOnly
		if !visibleInWorkspace && sameFile && permissionsSuffice {
			return nil
		}
		unmountCtx, cancel := boundedContext(ctx, mountTimeout)
		_, unmountErr := m.runtime.run(unmountCtx, "umount", m.runtime.name, mounted.hostPath+":"+source)
		cancel()
		if unmountErr != nil {
			return fmt.Errorf("refresh sbx bind source %q: %w", source, unmountErr)
		}
		delete(m.mounted, source)
	}
	if visibleInWorkspace {
		return nil
	}

	// sbx canonicalizes HOST before deriving its default target. Docker still
	// asks for source, so always pass that guest path explicitly.
	spec := resolved + ":" + source
	if request.readOnly {
		spec += ":ro"
	}
	mountCtx, cancel := boundedContext(ctx, mountTimeout)
	_, err = m.runtime.run(mountCtx, "mount", m.runtime.name, spec)
	cancel()
	if err != nil {
		return fmt.Errorf("make host bind source %q visible in sandbox %q: %w", source, m.runtime.name, err)
	}
	m.mounted[source] = mountedPath{hostPath: resolved, info: info, readOnly: request.readOnly}
	return nil
}

// workspacePathVisible follows symlinks as the guest would under the
// configured workspace target. A host path can resolve into the workspace
// while its original Docker path is absent in the guest, and an inner symlink
// can escape the mounted tree even when its final host target is inside it.
func workspacePathVisible(workspace, source string) bool {
	if !pathWithin(workspace, source) {
		return false
	}
	relative, err := filepath.Rel(workspace, source)
	if err != nil {
		return false
	}
	if relative == "." {
		return true
	}
	parts := strings.Split(relative, string(filepath.Separator))
	current := workspace
	links := 0
	for len(parts) > 0 {
		component := parts[0]
		parts = parts[1:]
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			return false
		}
		if info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		links++
		if links > 255 {
			return false
		}
		target, err := os.Readlink(current)
		if err != nil {
			return false
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(current), target)
		}
		target = filepath.Clean(target)
		if !pathWithin(workspace, target) {
			return false
		}
		remainder, err := filepath.Rel(workspace, target)
		if err != nil {
			return false
		}
		if remainder != "." {
			parts = append(strings.Split(remainder, string(filepath.Separator)), parts...)
		}
		current = workspace
	}
	return true
}

func pathWithin(root, path string) bool {
	if root == "" {
		return false
	}
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
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
			mergeRequestedBind(requests, mount.Source, mount.ReadOnly)
		}
	}
	return requests, nil
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
