package sbx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestProxyPreservesVersionQueryBodyAndDockerError(t *testing.T) {
	f := newFakeSandbox(t)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_, _ = w.Write([]byte("[]"))
		case "/events":
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/v1.45/containers/demo/archive":
			if r.URL.RawQuery != "path=%2Ftmp%2Ffile&x=a%2Bb" {
				t.Errorf("query changed: %s", r.URL.RawQuery)
			}
			body, _ := io.ReadAll(r.Body)
			if string(body) != "tar-binary-\x00-data" {
				t.Errorf("body changed: %q", body)
			}
			w.Header().Set("X-Docker-Detail", "preserved")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"message":"daemon error","unknown":{"field":42}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(backend.Close)
	u, _ := url.Parse(backend.URL)
	runtime := f.runtime(t)
	runtime.dockerURL = u
	proxy, err := newDockerProxy(context.Background(), runtime)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := proxy.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	client := &http.Client{Timeout: time.Second}
	req, err := http.NewRequest(http.MethodPut, "http://"+strings.TrimPrefix(proxy.Host(), "tcp://")+"/v1.45/containers/demo/archive?path=%2Ftmp%2Ffile&x=a%2Bb", strings.NewReader("tar-binary-\x00-data"))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict || resp.Header.Get("X-Docker-Detail") != "preserved" {
		t.Fatalf("response changed: %d %+v", resp.StatusCode, resp.Header)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil || string(payload["unknown"]) != `{"field":42}` {
		t.Fatalf("error body changed: %s (%v)", body, err)
	}
}

func TestProxyStreamsFirstLogChunkBeforeBackendFinishes(t *testing.T) {
	f := newFakeSandbox(t)
	finish := make(chan struct{})
	var finishOnce sync.Once
	closeFinish := func() { finishOnce.Do(func() { close(finish) }) }
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_, _ = w.Write([]byte("[]"))
		case "/events":
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/v1.45/containers/demo/logs":
			w.Header().Set("Content-Type", "application/vnd.docker.raw-stream")
			_, _ = w.Write([]byte("first"))
			w.(http.Flusher).Flush()
			select {
			case <-finish:
				_, _ = w.Write([]byte("second"))
			case <-r.Context().Done():
			}
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(backend.Close)
	u, _ := url.Parse(backend.URL)
	runtime := f.runtime(t)
	runtime.dockerURL = u
	proxy, err := newDockerProxy(context.Background(), runtime)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		closeFinish()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := proxy.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("http://" + strings.TrimPrefix(proxy.Host(), "tcp://") + "/v1.45/containers/demo/logs?follow=1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	first := make([]byte, 5)
	if _, err := io.ReadFull(resp.Body, first); err != nil {
		t.Fatalf("proxy buffered first chunk: %v", err)
	}
	if string(first) != "first" {
		t.Fatalf("first chunk: %q", first)
	}
	closeFinish()
	second, err := io.ReadAll(resp.Body)
	if err != nil || string(second) != "second" {
		t.Fatalf("second chunk: %q, %v", second, err)
	}
}

func TestProxyCreatePublishesBeforeForwardingAndPreservesUnknownFields(t *testing.T) {
	f := newFakeSandbox(t)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_, _ = w.Write([]byte("[]"))
		case "/events":
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/v1.45/containers/create":
			f.record("docker POST /containers/create")
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
			}
			var payload map[string]json.RawMessage
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Error(err)
			}
			if string(payload["FutureDockerField"]) != `{"some":"value"}` {
				t.Errorf("unknown create field lost: %s", body)
			}
			var host struct {
				PortBindings portBindings `json:"PortBindings"`
			}
			if err := json.Unmarshal(payload["HostConfig"], &host); err != nil {
				t.Error(err)
			}
			if host.PortBindings["80/tcp"][0].HostPort != "39136" {
				t.Errorf("binding not forwarded: %s", body)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"Id":"created-id","Warnings":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(backend.Close)
	u, _ := url.Parse(backend.URL)
	runtime := f.runtime(t)
	runtime.dockerURL = u
	proxy, err := newDockerProxy(context.Background(), runtime)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := proxy.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	client := &http.Client{Timeout: time.Second}
	requestBody := `{"Image":"nginx:alpine","FutureDockerField":{"some":"value"},"HostConfig":{"PortBindings":{"80/tcp":[{"HostIp":"127.0.0.1","HostPort":"39136"}]}}}`
	resp, err := client.Post("http://"+strings.TrimPrefix(proxy.Host(), "tcp://")+"/v1.45/containers/create?name=demo", "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated || !strings.Contains(string(responseBody), `"Id":"created-id"`) {
		t.Fatalf("create response: %d %s", resp.StatusCode, responseBody)
	}
	steps := f.snapshot()
	publication, creation := -1, -1
	for i, step := range steps {
		if strings.Contains(step, "--publish 127.0.0.1:39136:39136/tcp4") {
			publication = i
		}
		if step == "docker POST /containers/create" {
			creation = i
		}
	}
	if publication < 0 || creation <= publication {
		t.Fatalf("Docker create preceded port reservation: %v", steps)
	}
}
