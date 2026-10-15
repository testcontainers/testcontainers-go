package sbx

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

// This test deliberately has no availability skip once opted in: missing sbx,
// image pull failures, and provider regressions must fail the host/CI job.
func TestRealSandboxContainerLifecycle(t *testing.T) {
	if os.Getenv("TESTCONTAINERS_SBX_INTEGRATION") != "1" {
		t.Skip("real sandbox integration is opt-in")
	}
	if os.Getenv("TESTCONTAINERS_PROVIDER") != "sbx" {
		t.Fatal("TESTCONTAINERS_PROVIDER must be sbx for real integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{Image: "nginx:alpine", ExposedPorts: []string{"80/tcp"}}, Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Minute)
		defer cleanupCancel()
		if err := container.Terminate(cleanupCtx); err != nil {
			t.Error(err)
		}
	})
	if container.GetContainerID() == "" {
		t.Fatal("container has no stable ID")
	}
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "80/tcp")
	if err != nil {
		t.Fatal(err)
	}
	if host != "127.0.0.1" || port.Port() == "" {
		t.Fatalf("unexpected mapped endpoint: %s:%s", host, port.Port())
	}
	endpoint := "http://" + net.JoinHostPort(host, port.Port())
	client := &http.Client{Timeout: 3 * time.Second}
	deadline := time.Now().Add(30 * time.Second)
	for {
		resp, requestErr := client.Get(endpoint)
		if requestErr == nil {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK && strings.Contains(string(body), "nginx") {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("nginx endpoint never became ready: %v", requestErr)
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(250 * time.Millisecond):
		}
	}
	code, output, err := container.Exec(ctx, []string{"sh", "-c", "printf sbx-exec"})
	if err != nil || code != 0 {
		t.Fatalf("exec: exit=%d err=%v", code, err)
	}
	data, err := io.ReadAll(output)
	if err != nil || !strings.Contains(string(data), "sbx-exec") {
		t.Fatalf("exec output: %q err=%v", data, err)
	}
	if err := container.CopyToContainer(ctx, []byte("copied"), "/tmp/sbx-integration-copy", 0o644); err != nil {
		t.Fatal(err)
	}
	code, output, err = container.Exec(ctx, []string{"cat", "/tmp/sbx-integration-copy"})
	if err != nil || code != 0 {
		t.Fatalf("copy exec: exit=%d err=%v", code, err)
	}
	data, err = io.ReadAll(output)
	if err != nil || !strings.Contains(string(data), "copied") {
		t.Fatalf("copied content: %q err=%v", data, err)
	}
	logs, err := container.Logs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(io.Discard, logs)
	_ = logs.Close()
	if err != nil {
		t.Fatal(err)
	}
	stableID := container.GetContainerID()
	stablePort := port.Port()
	if err := container.Stop(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if err := container.Start(ctx); err != nil {
		t.Fatal(err)
	}
	restartedPort, err := container.MappedPort(ctx, "80/tcp")
	if err != nil {
		t.Fatal(err)
	}
	if container.GetContainerID() != stableID || restartedPort.Port() != stablePort {
		t.Fatalf("restart changed ID/port: %s/%s -> %s/%s", stableID, stablePort, container.GetContainerID(), restartedPort.Port())
	}

	// A physical host listener must make a fixed publication fail before the
	// Docker create reaches the inner daemon.
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	occupied := listener.Addr().(*net.TCPAddr).Port
	config, err := defaultProvider.Initialize(ctx)
	if err != nil {
		t.Fatal(err)
	}
	api := "http://" + strings.TrimPrefix(config.DockerHost, "tcp://")
	body := fmt.Sprintf(`{"Image":"nginx:alpine","HostConfig":{"PortBindings":{"80/tcp":[{"HostIp":"127.0.0.1","HostPort":"%d"}]}}}`, occupied)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, api+"/containers/create", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	responseBody, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if response.StatusCode < 400 {
		t.Fatalf("occupied physical host port %d accepted: HTTP %d %s", occupied, response.StatusCode, responseBody)
	}
}

func TestRealSandboxExternalSymlinkBind(t *testing.T) {
	if os.Getenv("TESTCONTAINERS_SBX_INTEGRATION") != "1" {
		t.Skip("real sandbox integration is opt-in")
	}
	if os.Getenv("TESTCONTAINERS_PROVIDER") != "sbx" {
		t.Fatal("TESTCONTAINERS_PROVIDER must be sbx for real integration")
	}
	configuration, err := readConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	outside, err := os.MkdirTemp("", "tc-sbx-bind-source-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(outside); err != nil {
			t.Error(err)
		}
	})
	if pathWithin(configuration.workspace, outside) {
		t.Fatalf("temporary bind source unexpectedly lies in workspace: %s", outside)
	}
	dataPath := filepath.Join(outside, "payload.txt")
	const content = "sbx-host-bind-content"
	if err := os.WriteFile(dataPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	aliasRoot, err := os.MkdirTemp("", "tc-sbx-bind-alias-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(aliasRoot); err != nil {
			t.Error(err)
		}
	})
	alias := filepath.Join(aliasRoot, "original-alias")
	if err := os.Symlink(outside, alias); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "nginx:alpine", Binds: []string{alias + ":/host-external:ro"},
		}, Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Minute)
		defer cleanupCancel()
		if err := container.Terminate(cleanupCtx); err != nil {
			t.Error(err)
		}
	})
	code, output, err := container.Exec(ctx, []string{"cat", "/host-external/payload.txt"})
	if err != nil || code != 0 {
		t.Fatalf("read bind via original symlink alias: exit=%d error=%v", code, err)
	}
	got, err := io.ReadAll(output)
	if err != nil || string(got) != content {
		t.Fatalf("host bind content: %q, error=%v", got, err)
	}
}
