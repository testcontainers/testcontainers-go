package trinogateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const backendStateHealthy = "HEALTHY"

const backendStatePollInterval = 500 * time.Millisecond

// Backend is a Trino cluster registered in the gateway.
type Backend struct {
	Name string `json:"name"`
	// ProxyTo is the URL of the Trino coordinator, as reachable from the gateway container.
	ProxyTo string `json:"proxyTo"`
	// ExternalURL is shown in the gateway UI instead of ProxyTo, if set.
	ExternalURL string `json:"externalUrl,omitempty"`
	Active      bool   `json:"active"`
	// RoutingGroup defaults to "adhoc".
	RoutingGroup string   `json:"routingGroup,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

// BackendState is the state of a backend, as reported by the gateway.
type BackendState struct {
	// TrinoStatus is "HEALTHY", "UNHEALTHY", "PENDING" or "UNKNOWN".
	TrinoStatus       string `json:"trinoStatus"`
	RunningQueryCount int    `json:"runningQueryCount"`
	QueuedQueryCount  int    `json:"queuedQueryCount"`
	NumWorkerNodes    int    `json:"numWorkerNodes"`
	ProxyTo           string `json:"proxyTo"`
	ExternalURL       string `json:"externalUrl"`
	RoutingGroup      string `json:"routingGroup"`
}

// Healthy reports whether the backend passed the gateway's health check.
func (s BackendState) Healthy() bool {
	return s.TrinoStatus == backendStateHealthy
}

// AddBackend registers a Trino cluster as an active backend. Queries are routed to it
// once it is healthy, see WaitForBackendHealthy.
func (c *Container) AddBackend(ctx context.Context, backend Backend) error {
	backend.Active = true

	body, err := json.Marshal(backend)
	if err != nil {
		return fmt.Errorf("marshal backend: %w", err)
	}

	return c.doRequest(ctx, http.MethodPost, "/gateway/backend/modify/add", "application/json", bytes.NewReader(body), nil)
}

// DeleteBackend removes a Trino cluster from the gateway.
func (c *Container) DeleteBackend(ctx context.Context, name string) error {
	return c.doRequest(ctx, http.MethodPost, "/gateway/backend/modify/delete", "text/plain", strings.NewReader(name), nil)
}

// ActivateBackend activates a Trino cluster, so that queries are routed to it.
func (c *Container) ActivateBackend(ctx context.Context, name string) error {
	return c.doRequest(ctx, http.MethodPost, "/gateway/backend/activate/"+url.PathEscape(name), "", nil, nil)
}

// DeactivateBackend deactivates a Trino cluster, so that no queries are routed to it.
func (c *Container) DeactivateBackend(ctx context.Context, name string) error {
	return c.doRequest(ctx, http.MethodPost, "/gateway/backend/deactivate/"+url.PathEscape(name), "", nil, nil)
}

// Backends returns all the Trino clusters registered in the gateway.
func (c *Container) Backends(ctx context.Context) ([]Backend, error) {
	var backends []Backend
	if err := c.doRequest(ctx, http.MethodGet, "/gateway/backend/all", "", nil, &backends); err != nil {
		return nil, err
	}

	return backends, nil
}

// ActiveBackends returns the active Trino clusters registered in the gateway.
func (c *Container) ActiveBackends(ctx context.Context) ([]Backend, error) {
	var backends []Backend
	if err := c.doRequest(ctx, http.MethodGet, "/gateway/backend/active", "", nil, &backends); err != nil {
		return nil, err
	}

	return backends, nil
}

// BackendState returns the state of a Trino cluster, as reported by the gateway.
func (c *Container) BackendState(ctx context.Context, name string) (BackendState, error) {
	var state BackendState
	if err := c.doRequest(ctx, http.MethodGet, "/api/public/backends/"+url.PathEscape(name)+"/state", "", nil, &state); err != nil {
		return BackendState{}, err
	}

	return state, nil
}

// WaitForBackendHealthy waits until the gateway reports the backend as healthy, or the context is done.
func (c *Container) WaitForBackendHealthy(ctx context.Context, name string) error {
	ticker := time.NewTicker(backendStatePollInterval)
	defer ticker.Stop()

	for {
		state, err := c.BackendState(ctx, name)
		if err != nil {
			return err
		}

		if state.Healthy() {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for backend %q healthy: last status %q: %w", name, state.TrinoStatus, ctx.Err())
		case <-ticker.C:
		}
	}
}

// doRequest calls the gateway API, decoding the JSON response into out if not nil.
func (c *Container) doRequest(ctx context.Context, method, path, contentType string, body io.Reader, out any) error {
	baseURL, err := c.ConnectionString(ctx)
	if err != nil {
		return fmt.Errorf("connection string: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s: unexpected status %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("%s %s: decode response: %w", method, path, err)
	}

	return nil
}
