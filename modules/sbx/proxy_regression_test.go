package sbx

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func regressionProxy(t *testing.T, f *fakeSandbox, handler http.Handler) (*dockerProxyServer, string) {
	t.Helper()
	backend := httptest.NewServer(handler)
	t.Cleanup(backend.Close)
	u, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	runtime := f.runtime(t)
	runtime.dockerURL = u
	server, err := newDockerProxy(context.Background(), runtime)
	if err != nil {
		t.Fatal(err)
	}
	proxy := server.(*dockerProxyServer)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := proxy.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	return proxy, "http://" + strings.TrimPrefix(proxy.Host(), "tcp://")
}

func TestRestartPolicyRejectedBeforeCreateOrUpdateSideEffects(t *testing.T) {
	f := newFakeSandbox(t)
	var mu sync.Mutex
	var forwarded []string
	var createdCount int
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_, _ = w.Write([]byte("[]"))
		case "/events":
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/v1.45/containers/create", "/v1.45/containers/existing/update":
			body, _ := io.ReadAll(r.Body)
			mu.Lock()
			forwarded = append(forwarded, r.URL.Path+" "+string(body))
			var createdID string
			if strings.HasSuffix(r.URL.Path, "/create") {
				createdCount++
				createdID = fmt.Sprintf("created-id-%d", createdCount)
			}
			mu.Unlock()
			if strings.HasSuffix(r.URL.Path, "/create") {
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]string{"Id": createdID})
			} else {
				_, _ = w.Write([]byte(`{"Warnings":[]}`))
			}
		default:
			http.NotFound(w, r)
		}
	})
	_, base := regressionProxy(t, f, backend)
	client := &http.Client{Timeout: time.Second}
	for _, route := range []string{"/v1.45/containers/create", "/v1.45/containers/existing/update"} {
		for _, policy := range []string{"always", "unless-stopped", "on-failure", "unexpected-policy"} {
			t.Run(route+"/"+policy, func(t *testing.T) {
				body := `{"Image":"nginx:alpine","FutureField":{"keep":true},"HostConfig":{"RestartPolicy":{"Name":"` + policy + `"},"Binds":["/definitely-missing-sbx-bind:/guest"],"PortBindings":{"80/tcp":[{"HostPort":"39207"}]}}}`
				if strings.HasSuffix(route, "/update") {
					body = `{"RestartPolicy":{"Name":"` + policy + `"},"FutureField":{"keep":true}}`
				}
				request, err := http.NewRequest(http.MethodPost, base+route, strings.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}
				response, err := client.Do(request)
				if err != nil {
					t.Fatal(err)
				}
				_, _ = io.Copy(io.Discard, response.Body)
				_ = response.Body.Close()
				if response.StatusCode < 400 {
					t.Fatalf("enabled or unknown restart policy accepted: %d", response.StatusCode)
				}
				mu.Lock()
				count := len(forwarded)
				mu.Unlock()
				if count != 0 {
					t.Fatalf("invalid policy reached inner Docker: %v", forwarded)
				}
				for _, step := range f.snapshot() {
					if strings.Contains(step, "--publish") || strings.Contains(step, "mount ") {
						t.Fatalf("invalid policy caused side effect: %v", f.snapshot())
					}
				}
			})
		}
	}
	for _, policy := range []string{"", "no"} {
		for _, route := range []string{"/v1.45/containers/create", "/v1.45/containers/existing/update"} {
			body := `{"Image":"nginx:alpine","FutureField":{"keep":true},"HostConfig":{"RestartPolicy":{"Name":"` + policy + `"},"FutureHostField":{"keep":true}}}`
			if strings.HasSuffix(route, "/update") {
				body = `{"RestartPolicy":{"Name":"` + policy + `"},"FutureField":{"keep":true}}`
			}
			request, err := http.NewRequest(http.MethodPost, base+route, strings.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			response, err := client.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode >= 400 {
				t.Fatalf("disabled policy rejected for %s: HTTP %d", route, response.StatusCode)
			}
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if len(forwarded) != 4 {
		t.Fatalf("accepted requests not forwarded: %v", forwarded)
	}
	for _, request := range forwarded {
		var payload map[string]json.RawMessage
		_, body, _ := strings.Cut(request, " ")
		if err := json.Unmarshal([]byte(body), &payload); err != nil {
			t.Fatal(err)
		}
		if string(payload["FutureField"]) != `{"keep":true}` {
			t.Fatalf("accepted payload lost unknown field: %s", body)
		}
		if strings.Contains(request, "/create") && !strings.Contains(string(payload["HostConfig"]), `"FutureHostField":{"keep":true}`) {
			t.Fatalf("accepted create lost unknown host field: %s", body)
		}
	}
}

func TestUnapprovedBindRejectedBeforeCreateAndPortSideEffects(t *testing.T) {
	f := newFakeSandbox(t)
	forwarded := make(chan struct{}, 1)
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_, _ = w.Write([]byte("[]"))
		case "/events":
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/v1.45/containers/create":
			forwarded <- struct{}{}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"Id":"unexpected-create"}`))
		default:
			http.NotFound(w, r)
		}
	})
	_, base := regressionProxy(t, f, backend)
	unapprovedRoot, err := os.MkdirTemp("", "tc-sbx-unapproved-bind-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(unapprovedRoot) })
	source := unapprovedRoot + "/data"
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name       string
		hostConfig string
	}{
		{name: "legacy binds", hostConfig: fmt.Sprintf(`"Binds":[%q]`, source+":/data:ro")},
		{name: "structured mounts", hostConfig: fmt.Sprintf(`"Mounts":[{"Type":"bind","Source":%q,"Target":"/data","ReadOnly":true}]`, source)},
	}
	client := &http.Client{Timeout: time.Second}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			body := `{"Image":"nginx:alpine","HostConfig":{` + test.hostConfig + `,"PortBindings":{"80/tcp":[{"HostIp":"127.0.0.1","HostPort":"39207"}]}}}`
			response, err := client.Post(base+"/v1.45/containers/create", "application/json", strings.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			responseBody, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if response.StatusCode != http.StatusBadRequest || !strings.Contains(string(responseBody), "outside") {
				t.Fatalf("unapproved bind response: HTTP %d %s", response.StatusCode, responseBody)
			}
			select {
			case <-forwarded:
				t.Fatal("unapproved bind reached inner Docker create")
			default:
			}
			for _, step := range f.snapshot() {
				if strings.Contains(step, "--publish") || strings.Contains(step, "mount ") {
					t.Fatalf("unapproved bind caused side effects: %v", f.snapshot())
				}
			}
		})
	}
}

func TestCreatePreparationDoesNotLeakMountsOrPorts(t *testing.T) {
	for _, mode := range []string{"port-conflict", "publish-failure", "mount-failure", "second-mount-failure", "unapproved-second-bind"} {
		t.Run(mode, func(t *testing.T) {
			f := newFakeSandbox(t)
			runtime := f.runtime(t)
			mountCalls := 0
			runtime.run = func(ctx context.Context, args ...string) ([]byte, error) {
				if args[0] == "mount" {
					mountCalls++
					f.record("sbx " + strings.Join(args, " "))
					if mode == "mount-failure" || (mode == "second-mount-failure" && mountCalls == 2) {
						return nil, errors.New("mount denied")
					}
					return nil, nil
				}
				if args[0] == "umount" {
					f.record("sbx " + strings.Join(args, " "))
					return nil, nil
				}
				return f.run(ctx, args...)
			}
			server, err := newDockerProxy(context.Background(), runtime)
			if err != nil {
				t.Fatal(err)
			}
			proxy := server.(*dockerProxyServer)
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				if err := proxy.Close(ctx); err != nil {
					t.Error(err)
				}
			})
			switch mode {
			case "port-conflict":
				if _, err := proxy.ports.prepareCreate(context.Background(), portBindings{"81/tcp": {{HostPort: "39219"}}}); err != nil {
					t.Fatal(err)
				}
			case "publish-failure":
				f.publishErr = errors.New("publish denied")
			}
			source := filepath.Join(canonicalTempDir(t), "data")
			if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
				t.Fatal(err)
			}
			binds := []string{source + ":/data:ro"}
			if mode == "second-mount-failure" {
				if err := os.WriteFile(source+"-second", []byte("data"), 0o600); err != nil {
					t.Fatal(err)
				}
				binds = append(binds, source+"-second:/second:ro")
			}
			if mode == "unapproved-second-bind" {
				binds = append(binds, "/zz-tcgo-unapproved:/denied:ro")
			}
			raw, err := json.Marshal(map[string]any{
				"Image": "nginx:alpine",
				"HostConfig": map[string]any{
					"Binds": binds, "PortBindings": portBindings{"80/tcp": {{HostPort: "39219"}}},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPost, "/containers/create", strings.NewReader(string(raw)))
			operation, err := proxy.prepareCreate(request, "")
			if err == nil || operation != nil {
				t.Fatalf("invalid create prepared: operation=%v err=%v", operation, err)
			}
			want := map[string]string{"port-conflict": "reserved", "publish-failure": "publish denied", "mount-failure": "mount denied", "second-mount-failure": "mount denied", "unapproved-second-bind": "outside"}[mode]
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("create failed for the wrong reason: %v", err)
			}
			if len(proxy.mounts.mounted) != 0 {
				t.Fatalf("failed preparation retained mounts: %v", proxy.mounts.mounted)
			}
			if mode != "mount-failure" && mode != "second-mount-failure" {
				for _, step := range f.snapshot() {
					if strings.Contains(step, "mount ") {
						t.Fatalf("failed preparation mounted a host path: %v", f.snapshot())
					}
				}
			}
			wantPorts := 0
			if mode == "port-conflict" {
				wantPorts = 1
			}
			f.mu.Lock()
			defer f.mu.Unlock()
			if len(f.ports) != wantPorts {
				t.Fatalf("failed preparation changed port ownership: %v", f.ports)
			}
			proxy.ports.mu.Lock()
			defer proxy.ports.mu.Unlock()
			if len(proxy.ports.creations) != wantPorts || len(proxy.ports.orphans) != 0 {
				t.Fatalf("failed preparation retained port leases: creations=%d orphans=%v", len(proxy.ports.creations), proxy.ports.orphans)
			}
		})
	}
}

func TestParentPathComponentRejectedBeforeCreate(t *testing.T) {
	f := newFakeSandbox(t)
	forwarded := make(chan struct{}, 1)
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_, _ = w.Write([]byte("[]"))
		case "/events":
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/v1.45/containers/create":
			forwarded <- struct{}{}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"Id":"unexpected-create"}`))
		default:
			http.NotFound(w, r)
		}
	})
	proxy, base := regressionProxy(t, f, backend)
	workspace := proxy.mounts.roots[0].canonical
	deep := workspace + "/deep"
	if err := os.MkdirAll(deep+"/nested", 0o700); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{workspace + "/secret", deep + "/secret"} {
		if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("deep/nested", workspace+"/link"); err != nil {
		t.Fatal(err)
	}
	source := workspace + "/link/../secret"
	cases := []struct {
		name       string
		hostConfig string
	}{
		{name: "legacy binds", hostConfig: fmt.Sprintf(`"Binds":[%q]`, source+":/data")},
		{name: "structured mounts", hostConfig: fmt.Sprintf(`"Mounts":[{"Type":"bind","Source":%q,"Target":"/data"}]`, source)},
	}
	client := &http.Client{Timeout: time.Second}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			body := `{"Image":"nginx:alpine","HostConfig":{` + test.hostConfig + `,"PortBindings":{"80/tcp":[{"HostPort":"39207"}]}}}`
			response, err := client.Post(base+"/v1.45/containers/create", "application/json", strings.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			responseBody, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if response.StatusCode != http.StatusBadRequest || !strings.Contains(string(responseBody), "must not contain") {
				t.Fatalf("parent path response: HTTP %d %s", response.StatusCode, responseBody)
			}
			select {
			case <-forwarded:
				t.Fatal("parent path reached inner Docker create")
			default:
			}
			for _, step := range f.snapshot() {
				if strings.Contains(step, "--publish") || strings.Contains(step, "mount ") {
					t.Fatalf("parent path caused side effects: %v", f.snapshot())
				}
			}
		})
	}
}

func TestLinkDeleteAndRenamedAlias404RetainRunningLease(t *testing.T) {
	for _, query := range []string{"link=1", "link=true", "link=TRUE", "link=TrUe", "link=t", ""} {
		t.Run(query, func(t *testing.T) {
			f := newFakeSandbox(t)
			backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/containers/json":
					_, _ = w.Write([]byte("[]"))
				case "/events":
					w.(http.Flusher).Flush()
					<-r.Context().Done()
				case "/containers/container-1/json":
					_, _ = w.Write([]byte(`{"Id":"container-1","Name":"/renamed","State":{"Running":true}}`))
				case "/v1.45/containers/old":
					if query == "" {
						http.Error(w, "old alias missing", http.StatusNotFound)
					} else {
						w.WriteHeader(http.StatusNoContent)
					}
				default:
					http.NotFound(w, r)
				}
			})
			proxy, base := regressionProxy(t, f, backend)
			prepared, err := proxy.ports.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "39208"}}})
			if err != nil {
				t.Fatal(err)
			}
			if err := proxy.ports.bindCreate(prepared.Token, "container-1", "old"); err != nil {
				t.Fatal(err)
			}
			request, err := http.NewRequest(http.MethodDelete, base+"/v1.45/containers/old?"+query, nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := (&http.Client{Timeout: time.Second}).Do(request)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusNotFound {
				t.Fatalf("unexpected delete response: %d", response.StatusCode)
			}
			f.mu.Lock()
			_, live := f.ports["127.0.0.1:39208:39208/tcp4"]
			f.mu.Unlock()
			if !live {
				t.Fatalf("link-only or alias 404 deleted running lease: %v", f.snapshot())
			}
			if _, err := proxy.ports.prepareCreate(context.Background(), portBindings{"81/tcp": {{HostPort: "39208"}}}); err == nil {
				t.Fatal("running lease reassigned after link/alias deletion")
			}
		})
	}
}

func TestCompletedUpgradesReleaseTrackedConnectionsAndCloseStopsActiveUpgrade(t *testing.T) {
	f := newFakeSandbox(t)
	backendClosed := make(chan struct{}, 8)
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_, _ = w.Write([]byte("[]"))
		case "/events":
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/v1.45/exec/demo/start", "/v1.45/containers/demo/attach":
			connection, buffer, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			defer connection.Close()
			_, _ = buffer.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n")
			if err := buffer.Flush(); err != nil {
				t.Error(err)
				return
			}
			_, _ = io.Copy(io.Discard, connection)
			backendClosed <- struct{}{}
		default:
			http.NotFound(w, r)
		}
	})
	proxy, _ := regressionProxy(t, f, backend)
	dial := func(path string) net.Conn {
		t.Helper()
		connection, err := net.DialTimeout("tcp", strings.TrimPrefix(proxy.Host(), "tcp://"), time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if err := connection.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
			t.Fatal(err)
		}
		if _, err := fmt.Fprintf(connection, "POST %s HTTP/1.1\r\nHost: docker\r\nConnection: Upgrade\r\nUpgrade: tcp\r\nContent-Length: 0\r\n\r\n", path); err != nil {
			t.Fatal(err)
		}
		response, err := http.ReadResponse(bufio.NewReader(connection), &http.Request{Method: http.MethodPost})
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != http.StatusSwitchingProtocols {
			t.Fatalf("upgrade failed: %d", response.StatusCode)
		}
		return connection
	}
	for _, path := range []string{"/v1.45/exec/demo/start", "/v1.45/containers/demo/attach", "/v1.45/exec/demo/start"} {
		connection := dial(path)
		if err := connection.Close(); err != nil {
			t.Fatal(err)
		}
		select {
		case <-backendClosed:
		case <-time.After(time.Second):
			t.Fatal("completed upgrade did not release backend")
		}
	}
	deadline := time.Now().Add(time.Second)
	for {
		proxy.connectionsMu.Lock()
		tracked := len(proxy.connections)
		proxy.connectionsMu.Unlock()
		if tracked == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d completed hijacks remain tracked", tracked)
		}
		runtime.Gosched()
	}
	active := dial("/v1.45/containers/demo/attach")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	if err := proxy.Close(ctx); err != nil {
		t.Fatal(err)
	}
	cancel()
	buf := make([]byte, 1)
	if _, err := active.Read(buf); err == nil {
		t.Fatal("active hijack survived proxy cleanup")
	}
	_ = active.Close()
}
