package couchbase

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	MGMT_PORT     = "8091"
	MGMT_SSL_PORT = "18091"

	VIEW_PORT     = "8092"
	VIEW_SSL_PORT = "18092"

	QUERY_PORT     = "8093"
	QUERY_SSL_PORT = "18093"

	SEARCH_PORT     = "8094"
	SEARCH_SSL_PORT = "18094"

	ANALYTICS_PORT     = "8095"
	ANALYTICS_SSL_PORT = "18095"

	EVENTING_PORT     = "8096"
	EVENTING_SSL_PORT = "18096"

	KV_PORT     = "11210"
	KV_SSL_PORT = "11207"
)

type clusterInit func(context.Context) error

// CouchbaseContainer represents the Couchbase container type used in the module
type CouchbaseContainer struct {
	testcontainers.Container
	config *Config
}

// StartContainer creates an instance of the Couchbase container type
func StartContainer(ctx context.Context, opts ...Option) (*CouchbaseContainer, error) {
	config := &Config{
		enabledServices: []service{kv, query, search, index},
		username:        "Administrator",
		password:        "password",
	}

	for _, opt := range opts {
		opt(config)
	}

	req := testcontainers.ContainerRequest{
		Image: "couchbase:6.5.1",
	}

	exposePorts(&req, config.enabledServices)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	couchbaseContainer := CouchbaseContainer{container, config}

	if err = couchbaseContainer.waitUntilAllNodesAreHealthy(ctx); err != nil {
		return nil, err
	}

	clusterInitFunc := []clusterInit{
		couchbaseContainer.waitUntilNodeIsOnline,
		couchbaseContainer.initializeIsEnterprise,
		couchbaseContainer.renameNode,
		couchbaseContainer.initializeServices,
		couchbaseContainer.setMemoryQuotas,
		couchbaseContainer.configureAdminUser,
		couchbaseContainer.configureExternalPorts,
	}

	if contains(config.enabledServices, index) {
		clusterInitFunc = append(clusterInitFunc, couchbaseContainer.configureIndexer)
	}

	for _, fn := range clusterInitFunc {
		if err = fn(ctx); err != nil {
			return nil, err
		}
	}

	return &couchbaseContainer, nil
}

func exposePorts(req *testcontainers.ContainerRequest, enabledServices []service) {
	req.ExposedPorts = append(req.ExposedPorts, MGMT_PORT, MGMT_SSL_PORT)

	for _, service := range enabledServices {
		req.ExposedPorts = append(req.ExposedPorts, service.ports...)
	}
}

func (c *CouchbaseContainer) waitUntilAllNodesAreHealthy(ctx context.Context) error {
	var waitStrategy []wait.Strategy

	waitStrategy = append(waitStrategy, wait.ForHTTP("/pools/default").
		WithPort(MGMT_PORT).
		WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusOK
		}).
		WithResponseMatcher(func(body io.Reader) bool {
			json, err := io.ReadAll(body)
			if err != nil {
				return false
			}
			status := gjson.Get(string(json), "nodes.0.status")
			if status.String() != "healthy" {
				return false
			}

			return true
		}))

	if contains(c.config.enabledServices, query) {
		waitStrategy = append(waitStrategy, wait.ForHTTP("/admin/ping").
			WithPort(QUERY_PORT).
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
		)
	}

	if contains(c.config.enabledServices, analytics) {
		waitStrategy = append(waitStrategy, wait.ForHTTP("/admin/ping").
			WithPort(ANALYTICS_PORT).
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}))
	}

	if contains(c.config.enabledServices, eventing) {
		waitStrategy = append(waitStrategy, wait.ForHTTP("/api/v1/config").
			WithPort(EVENTING_PORT).
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}))
	}

	return wait.ForAll(waitStrategy...).WaitUntilReady(ctx, c)
}

func (c *CouchbaseContainer) waitUntilNodeIsOnline(ctx context.Context) error {
	return wait.ForHTTP("/pools").
		WithPort(MGMT_PORT).
		WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusOK
		}).
		WaitUntilReady(ctx, c)
}

func (c *CouchbaseContainer) initializeIsEnterprise(ctx context.Context) error {
	response, err := c.doHttpRequest(ctx, MGMT_PORT, "/pools", http.MethodGet, nil, false)
	if err != nil {
		return err
	}

	c.config.isEnterprise = gjson.Get(string(response), "isEnterprise").Bool()

	if !c.config.isEnterprise {
		if contains(c.config.enabledServices, analytics) {
			return errors.New("the Analytics Service is only supported with the Enterprise version")
		}
		if contains(c.config.enabledServices, eventing) {
			return errors.New("the Eventing Service is only supported with the Enterprise version")
		}
	}

	return nil
}

func (c *CouchbaseContainer) renameNode(ctx context.Context) error {
	hostname, err := c.getInternalIPAddress(ctx)
	if err != nil {
		return err
	}

	body := map[string]string{
		"hostname": hostname,
	}

	_, err = c.doHttpRequest(ctx, MGMT_PORT, "/node/controller/rename", http.MethodPost, body, false)

	return err
}

func (c *CouchbaseContainer) initializeServices(ctx context.Context) error {
	body := map[string]string{
		"services": c.getEnabledServices(),
	}
	_, err := c.doHttpRequest(ctx, MGMT_PORT, "/node/controller/setupServices", http.MethodPost, body, false)

	return err
}

func (c *CouchbaseContainer) setMemoryQuotas(ctx context.Context) error {
	body := map[string]string{}

	for _, s := range c.config.enabledServices {
		if !s.hasQuota() {
			continue
		}

		quota := strconv.Itoa(s.minimumQuotaMb)
		if s.identifier == kv.identifier {
			body["memoryQuota"] = quota
		} else {
			body[s.identifier+"MemoryQuota"] = quota
		}
	}

	_, err := c.doHttpRequest(ctx, MGMT_PORT, "/pools/default", http.MethodPost, body, false)

	return err
}

func (c *CouchbaseContainer) configureAdminUser(ctx context.Context) error {
	body := map[string]string{
		"username": c.config.username,
		"password": c.config.password,
		"port":     "SAME",
	}

	_, err := c.doHttpRequest(ctx, MGMT_PORT, "/settings/web", http.MethodPost, body, false)

	return err
}

func (c *CouchbaseContainer) configureExternalPorts(ctx context.Context) error {
	host, _ := c.Host(ctx)
	mgmt, _ := c.MappedPort(ctx, MGMT_PORT)
	mgmtSSL, _ := c.MappedPort(ctx, MGMT_SSL_PORT)
	body := map[string]string{
		"hostname": host,
		"mgmt":     string(mgmt),
		"mgmtSSL":  string(mgmtSSL),
	}

	if contains(c.config.enabledServices, kv) {
		kv, _ := c.MappedPort(ctx, KV_PORT)
		kvSSL, _ := c.MappedPort(ctx, KV_SSL_PORT)
		capi, _ := c.MappedPort(ctx, VIEW_PORT)
		capiSSL, _ := c.MappedPort(ctx, VIEW_SSL_PORT)

		body["kv"] = string(kv)
		body["kvSSL"] = string(kvSSL)
		body["capi"] = string(capi)
		body["capiSSL"] = string(capiSSL)
	}

	if contains(c.config.enabledServices, query) {
		n1ql, _ := c.MappedPort(ctx, QUERY_PORT)
		n1qlSSL, _ := c.MappedPort(ctx, KV_SSL_PORT)

		body["n1ql"] = string(n1ql)
		body["n1qlSSL"] = string(n1qlSSL)
	}

	if contains(c.config.enabledServices, search) {
		fts, _ := c.MappedPort(ctx, SEARCH_PORT)
		ftsSSL, _ := c.MappedPort(ctx, SEARCH_SSL_PORT)

		body["fts"] = string(fts)
		body["ftsSSL"] = string(ftsSSL)
	}

	if contains(c.config.enabledServices, analytics) {
		cbas, _ := c.MappedPort(ctx, ANALYTICS_PORT)
		cbasSSL, _ := c.MappedPort(ctx, ANALYTICS_SSL_PORT)

		body["cbas"] = string(cbas)
		body["cbasSSL"] = string(cbasSSL)
	}

	if contains(c.config.enabledServices, eventing) {
		eventingAdminPort, _ := c.MappedPort(ctx, EVENTING_PORT)
		eventingSSL, _ := c.MappedPort(ctx, EVENTING_SSL_PORT)

		body["eventingAdminPort"] = string(eventingAdminPort)
		body["eventingSSL"] = string(eventingSSL)
	}

	_, err := c.doHttpRequest(ctx, MGMT_PORT, "/node/controller/setupAlternateAddresses/external", http.MethodPut, body, true)

	return err
}

func (c *CouchbaseContainer) configureIndexer(ctx context.Context) error {
	storageMode := "forestdb"
	if c.config.isEnterprise {
		storageMode = "memory_optimized"
	}

	body := map[string]string{
		"storageMode": storageMode,
	}

	_, err := c.doHttpRequest(ctx, MGMT_PORT, "/settings/indexes", http.MethodPost, body, true)

	return err
}

func (c *CouchbaseContainer) doHttpRequest(ctx context.Context, port, path, method string, body map[string]string, auth bool) ([]byte, error) {
	form := url.Values{}
	for k, v := range body {
		form.Set(k, v)
	}

	url, err := c.getUrl(ctx, port, path)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(method, url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if auth {
		request.SetBasicAuth(c.config.username, c.config.password)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return io.ReadAll(response.Body)
}

func (c *CouchbaseContainer) getUrl(ctx context.Context, port, path string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := c.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s%s", host, mappedPort, path), nil
}

func (c *CouchbaseContainer) getInternalIPAddress(ctx context.Context) (string, error) {
	networks, err := c.Networks(ctx)
	if err != nil {
		return "", err
	}

	return networks[0], nil
}

func (c *CouchbaseContainer) getEnabledServices() string {
	identifiers := make([]string, len(c.config.enabledServices))
	for _, v := range c.config.enabledServices {
		identifiers = append(identifiers, v.identifier)
	}

	return strings.Join(identifiers, ",")
}

func contains(services []service, service service) bool {
	for _, s := range services {
		if s.identifier == service.identifier {
			return true
		}
	}
	return false
}
