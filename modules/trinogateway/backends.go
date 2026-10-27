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

// backendStateHealthy is the status reported by the gateway for a backend
// whose health check succeeded.
const backendStateHealthy = "HEALTHY"

// backendStatePollInterval is the interval between two checks of the backend state.
const backendStatePollInterval = 500 * time.Millisecond

// Backend represents a Trino cluster registered in Trino Gateway.
type Backend struct {
	// Name is the unique name of the backend.
	Name string `json:"name"`
	// ProxyTo is the URL of the Trino coordinator, as reachable from the Trino Gateway container,
	// e.g. "http://trino:8080" when the Trino container has the "trino" network alias.
	ProxyTo string `json:"proxyTo"`
	// ExternalURL is an optional URL of the Trino coordinator to show in the gateway UI,
	// when it differs from ProxyTo. It defaults to ProxyTo.
	ExternalURL string `json:"externalUrl,omitempty"`
	// Active reports whether the backend receives queries. Backends are always
	// registered as active, use DeactivateBackend to deactivate them.
	Active bool `json:"active"`
	// RoutingGroup is the routing group of the backend. It defaults to "adhoc",
	// the group used for queries that do not specify one.
	RoutingGroup string `json:"routingGroup,omitempty"`
	// Tags are optional tags of the backend.
	Tags []string `json:"tags,omitempty"`
}

// BackendState represents the state of a backend as observed by the gateway's cluster monitor.
type BackendState struct {
	// TrinoStatus is the health status of the backend: "HEALTHY", "UNHEALTHY", "PENDING" or "UNKNOWN".
	TrinoStatus string `json:"trinoStatus"`
	// RunningQueryCount is the number of queries running in the backend.
	RunningQueryCount int `json:"runningQueryCount"`
	// QueuedQueryCount is the number of queries queued in the backend.
	QueuedQueryCount int `json:"queuedQueryCount"`
	// NumWorkerNodes is the number of worker nodes of the backend.
	NumWorkerNodes int `json:"numWorkerNodes"`
	// ProxyTo is the URL of the Trino coordinator.
	ProxyTo string `json:"proxyTo"`
	// ExternalURL is the external URL of the Trino coordinator.
	ExternalURL string `json:"externalUrl"`
	// RoutingGroup is the routing group of the backend.
	RoutingGroup string `json:"routingGroup"`
}

// Healthy reports whether the backend passed the gateway's health check.
func (s BackendState) Healthy() bool {
	return s.TrinoStatus == backendStateHealthy
}

// AddBackend registers a Trino cluster in the gateway. The backend is registered
// as active, regardless of the Active field.
// Queries are routed to the backend once the gateway's health check succeeds,
// use WaitForBackendHealthy to wait for it.
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

// BackendState returns the state of a Trino cluster, as observed by the gateway's cluster monitor.
func (c *Container) BackendState(ctx context.Context, name string) (BackendState, error) {
	var state BackendState
	if err := c.doRequest(ctx, http.MethodGet, "/api/public/backends/"+url.PathEscape(name)+"/state", "", nil, &state); err != nil {
		return BackendState{}, err
	}

	return state, nil
}

// WaitForBackendHealthy polls the state of a Trino cluster until the gateway reports it
// as healthy, or the context is done. The gateway checks the health of the backends
// every second with the default configuration.
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

// doRequest performs an HTTP request against the gateway API, decoding a JSON response into out when not nil.
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
