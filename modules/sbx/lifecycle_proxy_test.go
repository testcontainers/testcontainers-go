package sbx

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestUnmanagedContainerLifecyclePreservesDockerResponse(t *testing.T) {
	for _, state := range []string{"unmanaged", "missing", "inspect-error"} {
		for _, action := range []string{"start", "restart", "stop", "kill", "delete"} {
			t.Run(state+"/"+action, func(t *testing.T) {
				f := newFakeSandbox(t)
				forwarded := make(chan string, 1)
				backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/containers/json":
						_, _ = w.Write([]byte("[]"))
					case "/events":
						w.(http.Flusher).Flush()
						<-r.Context().Done()
					case "/containers/external/json":
						switch state {
						case "missing":
							http.NotFound(w, r)
						case "inspect-error":
							http.Error(w, "inspect unavailable", http.StatusInternalServerError)
						default:
							_ = json.NewEncoder(w).Encode(map[string]any{"Id": "external", "Config": map[string]any{"Labels": map[string]string{}}, "State": map[string]any{"Running": true}})
						}
					default:
						forwarded <- r.Method + " " + r.URL.RequestURI()
						if state == "missing" {
							w.Header().Set("Content-Type", "application/json")
							w.WriteHeader(http.StatusNotFound)
							_, _ = w.Write([]byte(`{"message":"No such container: external"}`))
						} else {
							w.WriteHeader(http.StatusNoContent)
						}
					}
				})
				_, base := regressionProxy(t, f, backend)
				method, path := http.MethodPost, "/v1.45/containers/external/"+action+"?t=1"
				if action == "delete" {
					method, path = http.MethodDelete, "/v1.45/containers/external?force=1"
				}
				request, err := http.NewRequest(method, base+path, nil)
				if err != nil {
					t.Fatal(err)
				}
				response, err := http.DefaultClient.Do(request)
				if err != nil {
					t.Fatal(err)
				}
				defer response.Body.Close()
				body, err := io.ReadAll(response.Body)
				if err != nil {
					t.Fatal(err)
				}
				wantStatus := http.StatusNoContent
				if state == "missing" {
					wantStatus = http.StatusNotFound
					if string(body) != `{"message":"No such container: external"}` {
						t.Fatalf("Docker error body changed: %s", body)
					}
				}
				if state == "inspect-error" {
					wantStatus = http.StatusBadGateway
				}
				if response.StatusCode != wantStatus {
					t.Fatalf("status=%d, want %d; body=%s", response.StatusCode, wantStatus, body)
				}
				wantForwarded := state != "inspect-error" || (action != "start" && action != "restart")
				select {
				case got := <-forwarded:
					if !wantForwarded || got != method+" "+path {
						t.Fatalf("unexpected forwarded request: %s", got)
					}
				default:
					if wantForwarded {
						t.Fatal("lifecycle request was not forwarded")
					}
				}
				for _, step := range f.snapshot() {
					if strings.Contains(step, "--publish") || strings.Contains(step, "--unpublish") {
						t.Fatalf("unmanaged lifecycle changed publications: %v", f.snapshot())
					}
				}
			})
		}
	}
}
