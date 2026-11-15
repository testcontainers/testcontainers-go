package sbx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCaseInsensitiveAndDuplicateJSONSettingsCannotBypassValidation(t *testing.T) {
	f := newFakeSandbox(t)
	forwarded := make(chan string, 32)
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_, _ = w.Write([]byte("[]"))
		case "/events":
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		default:
			forwarded <- r.URL.Path
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"Id":"unexpected-create"}`))
		}
	})
	_, base := regressionProxy(t, f, backend)
	cases := []struct {
		name   string
		update bool
		body   string
	}{
		{name: "lowercase-create-policy", body: `{"image":"nginx:alpine","hostconfig":{"restartpolicy":{"name":"always"}}}`},
		{name: "mixed-case-create-policy", body: `{"iMaGe":"nginx:alpine","hOsTcOnFiG":{"rEsTaRtPoLiCy":{"nAmE":"unless-stopped"}}}`},
		{name: "lowercase-update-policy", update: true, body: `{"restartpolicy":{"name":"on-failure"}}`},
		{name: "mixed-case-update-policy", update: true, body: `{"rEsTaRtPoLiCy":{"nAmE":"always"}}`},
		{name: "duplicate-exact-update-policy", update: true, body: `{"RestartPolicy":{"Name":"always"},"RestartPolicy":{}}`},
		{name: "duplicate-alias-update-policy", update: true, body: `{"RestartPolicy":{"Name":"always"},"restartpolicy":{}}`},
		{name: "duplicate-exact-update-name", update: true, body: `{"RestartPolicy":{"Name":"always","Name":"no"}}`},
		{name: "duplicate-alias-update-name", update: true, body: `{"RestartPolicy":{"Name":"always","name":"no"}}`},
		{name: "duplicate-exact-hostconfig", body: `{"Image":"nginx:alpine","HostConfig":{"RestartPolicy":{"Name":"always"}},"HostConfig":{}}`},
		{name: "duplicate-alias-hostconfig", body: `{"Image":"nginx:alpine","HostConfig":{"RestartPolicy":{"Name":"always"}},"hostconfig":{}}`},
		{name: "duplicate-exact-create-policy", body: `{"Image":"nginx:alpine","HostConfig":{"RestartPolicy":{"Name":"always"},"RestartPolicy":{}}}`},
		{name: "duplicate-alias-create-policy", body: `{"Image":"nginx:alpine","HostConfig":{"RestartPolicy":{"Name":"always"},"restartpolicy":{}}}`},
		{name: "duplicate-portbindings-alias", body: `{"Image":"nginx:alpine","HostConfig":{"PortBindings":{},"portbindings":{}}}`},
		{name: "duplicate-mount-type-alias", body: `{"Image":"nginx:alpine","HostConfig":{"Mounts":[{"Type":"bind","type":"volume","Source":"/missing-host-source","Target":"/guest"}]}}`},
		{name: "lowercase-networkmode", body: `{"image":"nginx:alpine","hostconfig":{"networkmode":"host"}}`},
		{name: "lowercase-reserved-label", body: fmt.Sprintf(`{"image":"nginx:alpine","labels":{%q:"caller-owned"}}`, portLeaseLabel)},
	}
	client := &http.Client{Timeout: time.Second}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			route := "/v1.45/containers/create"
			if test.update {
				route = "/v1.45/containers/existing/update"
			}
			response, err := client.Post(base+route, "application/json", strings.NewReader(test.body))
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if response.StatusCode < 400 {
				t.Fatalf("unsupported or ambiguous JSON accepted: HTTP %d %s", response.StatusCode, body)
			}
			select {
			case path := <-forwarded:
				t.Fatalf("invalid JSON settings reached inner Docker: %s", path)
			default:
			}
			for _, step := range f.snapshot() {
				if strings.Contains(step, "--publish") || strings.Contains(step, "sbx mount ") {
					t.Fatalf("invalid JSON settings caused side effects: %v", f.snapshot())
				}
			}
		})
	}
}

func TestCaseInsensitiveCreateStillSharesMountsAndReservesPorts(t *testing.T) {
	for _, mountField := range []string{"binds", "mounts"} {
		t.Run(mountField, func(t *testing.T) {
			f := newFakeSandbox(t)
			source := t.TempDir()
			canonical, err := filepath.EvalSymlinks(source)
			if err != nil {
				t.Fatal(err)
			}
			forwarded := make(chan []byte, 1)
			backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/containers/json":
					_, _ = w.Write([]byte("[]"))
				case "/events":
					w.(http.Flusher).Flush()
					<-r.Context().Done()
				case "/v1.45/containers/create":
					body, _ := io.ReadAll(r.Body)
					forwarded <- body
					f.record("docker create")
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write([]byte(`{"Id":"case-insensitive-create"}`))
				default:
					http.NotFound(w, r)
				}
			})
			proxy, base := regressionProxy(t, f, backend)
			proxy.mounts.runtime.run = func(ctx context.Context, args ...string) ([]byte, error) {
				if len(args) > 0 && args[0] == "mount" {
					f.record("sbx " + strings.Join(args, " "))
					return nil, nil
				}
				return f.run(ctx, args...)
			}
			mount := fmt.Sprintf(`"binds":[%q]`, source+":/guest:ro")
			if mountField == "mounts" {
				mount = fmt.Sprintf(`"mounts":[{"type":"bind","source":%q,"target":"/guest","readonly":true,"FutureMountField":{"keep":true}}]`, source)
			}
			body := fmt.Sprintf(`{"image":"nginx:alpine","FutureField":{"keep":true},"labels":{"CaseSensitive":"value"},"hostconfig":{"restartpolicy":{"name":"no"},"portbindings":{"80/tcp":[{"hostip":"127.0.0.1","hostport":"39211"}]},"FutureHostField":{"keep":true},%s}}`, mount)
			response, err := (&http.Client{Timeout: time.Second}).Post(base+"/v1.45/containers/create", "application/json", strings.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			responseBody, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if response.StatusCode != http.StatusCreated {
				t.Fatalf("lowercase create rejected: HTTP %d %s", response.StatusCode, responseBody)
			}
			var payload struct {
				Image       string
				FutureField map[string]bool
				Labels      map[string]string
				HostConfig  struct {
					RestartPolicy   struct{ Name string }
					PortBindings    portBindings
					FutureHostField map[string]bool
					Mounts          []struct{ FutureMountField map[string]bool }
				}
			}
			if err := json.Unmarshal(<-forwarded, &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Image != "nginx:alpine" || !payload.FutureField["keep"] || !payload.HostConfig.FutureHostField["keep"] || payload.Labels["CaseSensitive"] != "value" || payload.HostConfig.RestartPolicy.Name != "no" {
				t.Fatalf("accepted create lost fields: %+v", payload)
			}
			bindings := payload.HostConfig.PortBindings["80/tcp"]
			if len(bindings) != 1 || bindings[0].HostPort != "39211" {
				t.Fatalf("lowercase port bindings were not forwarded: %+v", bindings)
			}
			if mountField == "mounts" && (len(payload.HostConfig.Mounts) != 1 || !payload.HostConfig.Mounts[0].FutureMountField["keep"]) {
				t.Fatal("accepted structured mount lost unknown fields")
			}
			mountIndex, portIndex, createIndex := -1, -1, -1
			for i, step := range f.snapshot() {
				if step == "sbx mount fake-sandbox "+canonical+":"+source+":ro" {
					mountIndex = i
				}
				if strings.Contains(step, "--publish 127.0.0.1:39211:39211/tcp4") {
					portIndex = i
				}
				if step == "docker create" {
					createIndex = i
				}
			}
			if mountIndex < 0 || portIndex < 0 || createIndex <= mountIndex || createIndex <= portIndex {
				t.Fatalf("lowercase mounts/ports bypassed preparation: %v", f.snapshot())
			}
		})
	}
}

func TestCaseInsensitiveDisabledUpdatePreservesOriginalPayload(t *testing.T) {
	f := newFakeSandbox(t)
	forwarded := make(chan string, 1)
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_, _ = w.Write([]byte("[]"))
		case "/events":
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/v1.45/containers/existing/update":
			body, _ := io.ReadAll(r.Body)
			forwarded <- string(body)
			_, _ = w.Write([]byte(`{"Warnings":[]}`))
		default:
			http.NotFound(w, r)
		}
	})
	_, base := regressionProxy(t, f, backend)
	for _, policy := range []string{"", "no"} {
		body := fmt.Sprintf(`{"rEsTaRtPoLiCy":{"nAmE":%q},"FutureField":{"keep":true},"FutureField":{"keep":false}}`, policy)
		response, err := (&http.Client{Timeout: time.Second}).Post(base+"/v1.45/containers/existing/update", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		if response.StatusCode >= 400 {
			t.Fatalf("disabled mixed-case update rejected: HTTP %d", response.StatusCode)
		}
		if got := <-forwarded; got != body {
			t.Fatalf("update changed unrelated JSON semantics:\ngot %s\nwant %s", got, body)
		}
	}
}
