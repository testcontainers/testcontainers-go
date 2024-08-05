package couchbase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/go-connections/nat"
	"github.com/tidwall/gjson"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// containerPorts {

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

	// }
)

// initialServices is the list of services that are enabled by default
var initialServices = []Service{kv, query, search, index}

type clusterInit func(context.Context) error

// CouchbaseContainer represents the Couchbase container type used in the module
type CouchbaseContainer struct {
	testcontainers.Container
	config *Config
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Couchbase container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CouchbaseContainer, error) {
	return Run(ctx, "couchbase:6.5.1", opts...)
}

// Run creates an instance of the Couchbase container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*CouchbaseContainer, error) {
	config := &Config{
		enabledServices:  make([]Service, 0),
		username:         "Administrator",
		password:         "password",
		indexStorageMode: MemoryOptimized,
	}

	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{MGMT_PORT + "/tcp", MGMT_SSL_PORT + "/tcp"},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, srv := range initialServices {
		opts = append(opts, withService(srv))
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}

		// transfer options to the config

		if bucketCustomizer, ok := opt.(bucketCustomizer); ok {
			// If the option is a bucketCustomizer, we need to add the buckets to the request
			config.buckets = append(config.buckets, bucketCustomizer.buckets...)
		} else if serviceCustomizer, ok := opt.(serviceCustomizer); ok {
			// If the option is a serviceCustomizer, we need to append the service
			config.enabledServices = append(config.enabledServices, serviceCustomizer.enabledService)
		} else if indexStorageCustomizer, ok := opt.(indexStorageCustomizer); ok {
			// If the option is a indexStorageCustomizer, we need to set the index storage mode
			config.indexStorageMode = indexStorageCustomizer.mode
		} else if credentialsCustomizer, ok := opt.(credentialsCustomizer); ok {
			// If the option is a credentialsCustomizer, we need to set the credentials
			config.username = credentialsCustomizer.username
			config.password = credentialsCustomizer.password

			if len(credentialsCustomizer.password) < 6 {
				return nil, errors.New("admin password must be at most 6 characters long")
			}
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var couchbaseContainer *CouchbaseContainer
	if container != nil {
		couchbaseContainer = &CouchbaseContainer{container, config}
	}
	if err != nil {
		return couchbaseContainer, err
	}

	if err = couchbaseContainer.initCluster(ctx); err != nil {
		return couchbaseContainer, fmt.Errorf("init cluster: %w", err)
	}

	if err = couchbaseContainer.createBuckets(ctx); err != nil {
		return couchbaseContainer, fmt.Errorf("create buckets: %w", err)
	}

	return couchbaseContainer, nil
}

// StartContainer creates an instance of the Couchbase container type
// Deprecated: use RunContainer instead
func StartContainer(ctx context.Context, opts ...Option) (*CouchbaseContainer, error) {
	config := &Config{
		enabledServices:  []Service{kv, query, search, index},
		username:         "Administrator",
		password:         "password",
		imageName:        "couchbase:6.5.1",
		indexStorageMode: MemoryOptimized,
	}

	for _, opt := range opts {
		opt(config)
	}

	customizers := []testcontainers.ContainerCustomizer{
		WithAdminCredentials(config.username, config.password),
		WithIndexStorage(config.indexStorageMode),
		WithBuckets(config.buckets...),
	}

	for _, srv := range config.enabledServices {
		customizers = append(customizers, withService(srv))
	}

	return Run(ctx, config.imageName, customizers...)
}

// ConnectionString returns the connection string to connect to the Couchbase container instance.
// It returns a string with the format couchbase://<host>:<port>
func (c *CouchbaseContainer) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, KV_PORT)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("couchbase://%s:%d", host, port.Int()), nil
}

// Username returns the username of the Couchbase administrator.
func (c *CouchbaseContainer) Username() string {
	return c.config.username
}

// Password returns the password of the Couchbase administrator.
func (c *CouchbaseContainer) Password() string {
	return c.config.password
}

func (c *CouchbaseContainer) initCluster(ctx context.Context) error {
	clusterInitFunc := []clusterInit{
		c.waitUntilNodeIsOnline,
		c.initializeIsEnterprise,
		c.renameNode,
		c.initializeServices,
		c.setMemoryQuotas,
		c.configureAdminUser,
		c.configureExternalPorts,
	}

	if contains(c.config.enabledServices, index) {
		clusterInitFunc = append(clusterInitFunc, c.configureIndexer)
	}

	clusterInitFunc = append(clusterInitFunc, c.waitUntilAllNodesAreHealthy)

	for _, fn := range clusterInitFunc {
		if err := fn(ctx); err != nil {
			return err
		}
	}

	return nil
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
		"mgmt":     mgmt.Port(),
		"mgmtSSL":  mgmtSSL.Port(),
	}

	if contains(c.config.enabledServices, kv) {
		kv, _ := c.MappedPort(ctx, KV_PORT)
		kvSSL, _ := c.MappedPort(ctx, KV_SSL_PORT)
		capi, _ := c.MappedPort(ctx, VIEW_PORT)
		capiSSL, _ := c.MappedPort(ctx, VIEW_SSL_PORT)

		body["kv"] = kv.Port()
		body["kvSSL"] = kvSSL.Port()
		body["capi"] = capi.Port()
		body["capiSSL"] = capiSSL.Port()
	}

	if contains(c.config.enabledServices, query) {
		n1ql, _ := c.MappedPort(ctx, QUERY_PORT)
		n1qlSSL, _ := c.MappedPort(ctx, QUERY_SSL_PORT)

		body["n1ql"] = n1ql.Port()
		body["n1qlSSL"] = n1qlSSL.Port()
	}

	if contains(c.config.enabledServices, search) {
		fts, _ := c.MappedPort(ctx, SEARCH_PORT)
		ftsSSL, _ := c.MappedPort(ctx, SEARCH_SSL_PORT)

		body["fts"] = fts.Port()
		body["ftsSSL"] = ftsSSL.Port()
	}

	if contains(c.config.enabledServices, analytics) {
		cbas, _ := c.MappedPort(ctx, ANALYTICS_PORT)
		cbasSSL, _ := c.MappedPort(ctx, ANALYTICS_SSL_PORT)

		body["cbas"] = cbas.Port()
		body["cbasSSL"] = cbasSSL.Port()
	}

	if contains(c.config.enabledServices, eventing) {
		eventingAdminPort, _ := c.MappedPort(ctx, EVENTING_PORT)
		eventingSSL, _ := c.MappedPort(ctx, EVENTING_SSL_PORT)

		body["eventingAdminPort"] = eventingAdminPort.Port()
		body["eventingSSL"] = eventingSSL.Port()
	}

	_, err := c.doHttpRequest(ctx, MGMT_PORT, "/node/controller/setupAlternateAddresses/external", http.MethodPut, body, true)

	return err
}

func (c *CouchbaseContainer) configureIndexer(ctx context.Context) error {
	if c.config.isEnterprise {
		if c.config.indexStorageMode == ForestDB {
			c.config.indexStorageMode = MemoryOptimized
		}
	} else {
		c.config.indexStorageMode = ForestDB
	}

	body := map[string]string{
		"storageMode": string(c.config.indexStorageMode),
	}

	_, err := c.doHttpRequest(ctx, MGMT_PORT, "/settings/indexes", http.MethodPost, body, true)

	return err
}

func (c *CouchbaseContainer) waitUntilAllNodesAreHealthy(ctx context.Context) error {
	var waitStrategy []wait.Strategy

	waitStrategy = append(waitStrategy, wait.ForHTTP("/pools/default").
		WithPort(MGMT_PORT).
		WithBasicAuth(c.config.username, c.config.password).
		WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusOK
		}).
		WithResponseMatcher(func(body io.Reader) bool {
			response, err := io.ReadAll(body)
			if err != nil {
				return false
			}
			status := gjson.Get(string(response), "nodes.0.status")

			return status.String() == "healthy"
		}))

	if contains(c.config.enabledServices, query) {
		waitStrategy = append(waitStrategy, wait.ForHTTP("/admin/ping").
			WithPort(QUERY_PORT).
			WithBasicAuth(c.config.username, c.config.password).
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
		)
	}

	if contains(c.config.enabledServices, analytics) {
		waitStrategy = append(waitStrategy, wait.ForHTTP("/admin/ping").
			WithPort(ANALYTICS_PORT).
			WithBasicAuth(c.config.username, c.config.password).
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}))
	}

	if contains(c.config.enabledServices, eventing) {
		waitStrategy = append(waitStrategy, wait.ForHTTP("/api/v1/config").
			WithPort(EVENTING_PORT).
			WithBasicAuth(c.config.username, c.config.password).
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}))
	}

	return wait.ForAll(waitStrategy...).WaitUntilReady(ctx, c)
}

func (c *CouchbaseContainer) createBuckets(ctx context.Context) error {
	for _, bucket := range c.config.buckets {
		err := c.createBucket(ctx, bucket)
		if err != nil {
			return err
		}

		err = c.waitForAllServicesEnabled(ctx, bucket)
		if err != nil {
			return err
		}

		if contains(c.config.enabledServices, query) {
			err = c.isQueryKeyspacePresent(ctx, bucket)
			if err != nil {
				return err
			}
		}

		if bucket.queryPrimaryIndex {
			if !contains(c.config.enabledServices, query) {
				return fmt.Errorf("primary index creation for bucket %s ignored, since QUERY service is not present", bucket.name)
			}

			err = c.createPrimaryIndex(ctx, bucket)
			if err != nil {
				return err
			}

			err = c.isPrimaryIndexOnline(ctx, bucket)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func (c *CouchbaseContainer) isPrimaryIndexOnline(ctx context.Context, bucket bucket) error {
	body := map[string]string{
		"statement": "SELECT count(*) > 0 AS online FROM system:indexes where keyspace_id = \"" +
			bucket.name +
			"\" and is_primary = true and state = \"online\"",
	}

	err := backoff.Retry(func() error {
		response, err := c.doHttpRequest(ctx, QUERY_PORT, "/query/service", http.MethodPost, body, true)
		if err != nil {
			return err
		}

		online := gjson.Get(string(response), "results.0.online").Bool()
		if !online {
			return errors.New("primary index state is not online")
		}

		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))

	return err
}

func (c *CouchbaseContainer) createPrimaryIndex(ctx context.Context, bucket bucket) error {
	body := map[string]string{
		"statement": "CREATE PRIMARY INDEX on `" + bucket.name + "`",
	}
	err := backoff.Retry(func() error {
		response, err := c.doHttpRequest(ctx, QUERY_PORT, "/query/service", http.MethodPost, body, true)
		firstError := gjson.Get(string(response), "errors.0.code").Int()
		if firstError != 0 {
			return errors.New("index creation failed")
		}
		return err
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
	return err
}

func (c *CouchbaseContainer) isQueryKeyspacePresent(ctx context.Context, bucket bucket) error {
	body := map[string]string{
		"statement": "SELECT COUNT(*) > 0 as present FROM system:keyspaces WHERE name = \"" + bucket.name + "\"",
	}

	err := backoff.Retry(func() error {
		response, err := c.doHttpRequest(ctx, QUERY_PORT, "/query/service", http.MethodPost, body, true)
		if err != nil {
			return err
		}
		present := gjson.Get(string(response), "results.0.present").Bool()
		if !present {
			return errors.New("query namespace is not present")
		}

		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))

	return err
}

func (c *CouchbaseContainer) waitForAllServicesEnabled(ctx context.Context, bucket bucket) error {
	err := wait.ForHTTP("/pools/default/b/"+bucket.name).
		WithPort(MGMT_PORT).
		WithBasicAuth(c.config.username, c.config.password).
		WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusOK
		}).
		WithResponseMatcher(func(body io.Reader) bool {
			response, err := io.ReadAll(body)
			if err != nil {
				return false
			}
			return c.checkAllServicesEnabled(response)
		}).
		WaitUntilReady(ctx, c)

	return err
}

func (c *CouchbaseContainer) createBucket(ctx context.Context, bucket bucket) error {
	flushEnabled := "0"
	if bucket.flushEnabled {
		flushEnabled = "1"
	}
	body := map[string]string{
		"name":          bucket.name,
		"ramQuotaMB":    strconv.Itoa(bucket.quota),
		"flushEnabled":  flushEnabled,
		"replicaNumber": strconv.Itoa(bucket.numReplicas),
	}

	_, err := c.doHttpRequest(ctx, MGMT_PORT, "/pools/default/buckets", http.MethodPost, body, true)

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

	var bytes []byte
	// retry with backoff
	backoffErr := backoff.Retry(func() error {
		request, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(form.Encode()))
		if err != nil {
			return err
		}

		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		if auth {
			request.SetBasicAuth(c.config.username, c.config.password)
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		bytes, err = io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))

	if backoffErr != nil {
		return nil, backoffErr
	}

	return bytes, nil
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

	return fmt.Sprintf("http://%s:%d%s", host, mappedPort.Int(), path), nil
}

func (c *CouchbaseContainer) getInternalIPAddress(ctx context.Context) (string, error) {
	networks, err := c.ContainerIP(ctx)
	if err != nil {
		return "", err
	}

	return networks, nil
}

func (c *CouchbaseContainer) getEnabledServices() string {
	identifiers := make([]string, len(c.config.enabledServices))
	for i, v := range c.config.enabledServices {
		identifiers[i] = v.identifier
	}

	return strings.Join(identifiers, ",")
}

func (c *CouchbaseContainer) checkAllServicesEnabled(rawConfig []byte) bool {
	nodeExt := gjson.Get(string(rawConfig), "nodesExt")
	if !nodeExt.Exists() {
		return false
	}

	for _, node := range nodeExt.Array() {
		services := node.Map()["services"]
		if !services.Exists() {
			return false
		}

		for _, s := range c.config.enabledServices {
			found := false
			for serviceName := range services.Map() {
				if strings.HasPrefix(serviceName, s.identifier) {
					found = true
				}
			}

			if !found {
				return false
			}
		}
	}

	return true
}

type serviceCustomizer struct {
	enabledService Service
}

func (c serviceCustomizer) Customize(req *testcontainers.GenericContainerRequest) error {
	for _, port := range c.enabledService.ports {
		req.ExposedPorts = append(req.ExposedPorts, port+"/tcp")
	}

	return nil
}

// withService creates a serviceCustomizer for the given service.
// It's private to prevent users from creating other services than the Analytics and Eventing services.
func withService(service Service) serviceCustomizer {
	return serviceCustomizer{
		enabledService: service,
	}
}

// WithServiceAnalytics enables the Analytics service.
func WithServiceAnalytics() serviceCustomizer {
	return withService(analytics)
}

// WithServiceEventing enables the Eventing service.
func WithServiceEventing() serviceCustomizer {
	return withService(eventing)
}

func contains(services []Service, service Service) bool {
	for _, s := range services {
		if s.identifier == service.identifier {
			return true
		}
	}
	return false
}
