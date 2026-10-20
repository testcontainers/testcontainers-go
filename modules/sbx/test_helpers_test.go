package sbx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// fakeSandbox records the externally visible order of CLI and Docker calls.
type fakeSandbox struct {
	mu           sync.Mutex
	steps        []string
	ports        map[string]portPublication
	containers   map[string]bool
	publishErr   error
	unpublishErr error
	list         []dockerContainer
	server       *httptest.Server
}

func newFakeSandbox(t *testing.T) *fakeSandbox {
	t.Helper()
	f := &fakeSandbox{ports: map[string]portPublication{}, containers: map[string]bool{}}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.record("docker " + r.Method + " " + r.URL.Path)
		switch {
		case r.URL.Path == "/_ping":
			_, _ = w.Write([]byte("OK"))
		case r.URL.Path == "/containers/json":
			f.mu.Lock()
			rows := append([]dockerContainer(nil), f.list...)
			f.mu.Unlock()
			out := make([]map[string]any, 0, len(rows))
			for _, c := range rows {
				out = append(out, map[string]any{"Id": c.ID, "Names": []string{c.Name}, "Labels": c.Labels, "Ports": c.Ports})
			}
			_ = json.NewEncoder(w).Encode(out)
		case strings.HasPrefix(r.URL.Path, "/containers/") && strings.HasSuffix(r.URL.Path, "/json"):
			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/containers/"), "/json")
			f.mu.Lock()
			running, ok := f.containers[id]
			f.mu.Unlock()
			if !ok {
				http.NotFound(w, r)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"Id": id, "Name": "/sample", "State": map[string]any{"Running": running}})
		case r.URL.Path == "/events":
			w.Header().Set("Content-Type", "application/json")
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeSandbox) record(s string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.steps = append(f.steps, s)
}

func (f *fakeSandbox) snapshot() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.steps...)
}

func (f *fakeSandbox) runtime(t *testing.T) sandboxRuntime {
	t.Helper()
	u, err := url.Parse(f.server.URL)
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	root, err := configureBindRoot(workspace, true)
	if err != nil {
		t.Fatal(err)
	}
	testRoot, err := configureBindRoot(filepath.Dir(workspace), false)
	if err != nil {
		t.Fatal(err)
	}
	return sandboxRuntime{name: "fake-sandbox", dockerURL: u, run: f.run, workspace: workspace, bindRoots: []configuredBindRoot{root, testRoot}}
}

func (f *fakeSandbox) run(ctx context.Context, args ...string) ([]byte, error) {
	f.record("sbx " + strings.Join(args, " "))
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(args) < 2 || args[0] != "ports" {
		return nil, fmt.Errorf("unexpected command: %v", args)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(args) > 2 && args[2] == "--publish" {
		if f.publishErr != nil {
			return nil, f.publishErr
		}
		var p portPublication
		// CLI reports the exact tuple it accepted.
		if _, err := fmt.Sscanf(args[3], "127.0.0.1:%d:%d/", &p.HostPort, &p.SandboxPort); err != nil {
			return nil, err
		}
		p.HostIP = "127.0.0.1"
		p.Protocol = strings.TrimPrefix(args[3][strings.LastIndex(args[3], "/"):], "/")
		f.ports[args[3]] = p
		return json.Marshal([]portPublication{p})
	}
	if len(args) > 2 && args[2] == "--unpublish" {
		if f.unpublishErr != nil {
			return nil, f.unpublishErr
		}
		delete(f.ports, args[3])
		return []byte("[]"), nil
	}
	rows := make([]portPublication, 0, len(f.ports))
	for _, p := range f.ports {
		rows = append(rows, p)
	}
	return json.Marshal(rows)
}
