package testcontainers

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/cpuguy83/dockercfg"
	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/client"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// defaultRegistryFn is variable overwritten in tests to check for behaviour with different default values.
var defaultRegistryFn = defaultRegistry

// getRegistryCredentials is a variable overwritten in tests to mock the dockercfg.GetRegistryCredentials function.
var getRegistryCredentials = dockercfg.GetRegistryCredentials

// DockerImageAuth returns the auth config for the given Docker image, extracting first its Docker registry.
// Finally, it will use the credential helpers to extract the information from the docker config file
// for that registry, if it exists.
func DockerImageAuth(ctx context.Context, image string) (string, registry.AuthConfig, error) {
	defaultRegistry := defaultRegistryFn(ctx)
	reg := core.ExtractRegistry(image, defaultRegistry)
	lookupReg := reg
	if strings.EqualFold(reg, "docker.io") ||
		strings.EqualFold(reg, "registry.hub.docker.com") ||
		strings.EqualFold(reg, "registry-1.docker.io") {
		lookupReg = defaultRegistry
	}

	configs, err := getDockerAuthConfigs(lookupReg)
	if err != nil {
		return reg, registry.AuthConfig{}, err
	}

	return dockerImageAuth(ctx, image, configs)
}

// dockerImageAuth returns the auth config for the given Docker image.
func dockerImageAuth(ctx context.Context, image string, configs map[string]registry.AuthConfig) (string, registry.AuthConfig, error) {
	defaultRegistry := defaultRegistryFn(ctx)
	reg := core.ExtractRegistry(image, defaultRegistry)

	// Normalize Docker Hub aliases for credential lookup
	if strings.EqualFold(reg, "docker.io") ||
		strings.EqualFold(reg, "registry.hub.docker.com") ||
		strings.EqualFold(reg, "registry-1.docker.io") {
		reg = defaultRegistry // This is https://index.docker.io/v1/
	}

	if cfg, ok := getRegistryAuth(reg, configs); ok {
		return reg, cfg, nil
	}

	return reg, registry.AuthConfig{}, dockercfg.ErrCredentialsNotFound
}

// normalizeRegistryHost extracts and normalizes the host portion of a registry key or URL for matching.
func normalizeRegistryHost(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	keyURL, err := url.Parse(s)
	if err == nil && keyURL.Host != "" {
		return strings.ToLower(keyURL.Host)
	}

	// Fallback when url.Parse treats the string as a relative path (e.g. without scheme)
	host := s
	if keyURL != nil && keyURL.Path != "" {
		host = keyURL.Path
	}

	// Strip any scheme case-insensitively if url.Parse didn't populate Host
	if idx := strings.Index(strings.ToLower(host), "://"); idx != -1 {
		host = host[idx+3:]
	}

	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}

	return strings.ToLower(host)
}

// isDockerHubRegistry checks if a normalized registry host is one of Docker Hub's known domains.
func isDockerHubRegistry(s string) bool {
	return s == "docker.io" || s == "index.docker.io" || s == "registry-1.docker.io" || s == "registry.hub.docker.com"
}

func getRegistryAuth(reg string, cfgs map[string]registry.AuthConfig) (registry.AuthConfig, bool) {
	if cfg, ok := cfgs[reg]; ok {
		return cfg, true
	}

	normReg := normalizeRegistryHost(reg)

	// First pass: fallback match using exact normalized host
	for k, cfg := range cfgs {
		if normalizeRegistryHost(k) == normReg {
			return cfg, true
		}
	}

	// Second pass: if target is Docker Hub, fall back to any Docker Hub alias
	if isDockerHubRegistry(normReg) {
		for k, cfg := range cfgs {
			if isDockerHubRegistry(normalizeRegistryHost(k)) {
				return cfg, true
			}
		}
	}

	return registry.AuthConfig{}, false
}

// defaultRegistry returns the default registry to use when pulling images
// It will use the docker daemon to get the default registry, returning "https://index.docker.io/v1/" if
// it fails to get the information from the daemon
func defaultRegistry(ctx context.Context) string {
	apiClient, err := NewDockerClientWithOpts(ctx)
	if err != nil {
		return core.IndexDockerIO
	}
	defer apiClient.Close()

	info, err := apiClient.Info(ctx, client.InfoOptions{})
	if err != nil {
		return core.IndexDockerIO
	}

	return info.Info.IndexServerAddress
}

// authConfigResult is a result looking up auth details for key.
type authConfigResult struct {
	key string
	cfg registry.AuthConfig
	err error
}

// credentialsCache is a cache for registry credentials.
type credentialsCache struct {
	entries map[string]credentials
	mtx     sync.RWMutex
}

// credentials represents the username and password for a registry.
type credentials struct {
	username string
	password string
}

var creds = &credentialsCache{entries: map[string]credentials{}}

// AuthConfig updates the details in authConfig for the given hostname
// as determined by the details in configKey.
func (c *credentialsCache) AuthConfig(hostname, configKey string, authConfig *registry.AuthConfig) error {
	u, p, err := creds.get(hostname, configKey)
	if err != nil {
		return err
	}

	if u != "" {
		authConfig.Username = u
		authConfig.Password = p
	} else {
		authConfig.IdentityToken = p
	}

	return nil
}

// get returns the username and password for the given hostname
// as determined by the details in configPath.
// If the username is empty, the password is an identity token.
func (c *credentialsCache) get(hostname, configKey string) (string, string, error) {
	key := configKey + ":" + hostname
	c.mtx.RLock()
	entry, ok := c.entries[key]
	c.mtx.RUnlock()

	if ok {
		return entry.username, entry.password, nil
	}

	// No entry found, request and cache.
	user, password, err := getRegistryCredentials(hostname)
	if err != nil {
		return "", "", fmt.Errorf("getting credentials for %s: %w", hostname, err)
	}

	c.mtx.Lock()
	c.entries[key] = credentials{username: user, password: password}
	c.mtx.Unlock()

	return user, password, nil
}

// configKey returns a key to use for caching credentials based on
// the contents of the currently active config.
func configKey(cfg *dockercfg.Config) (string, error) {
	h := md5.New()
	if err := json.NewEncoder(h).Encode(cfg); err != nil {
		return "", fmt.Errorf("encode config: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// matchesRegistry checks if the given config key matches the target registry.
func matchesRegistry(key, target string) bool {
	if strings.EqualFold(key, target) {
		return true
	}

	normKey := normalizeRegistryHost(key)
	normTarget := normalizeRegistryHost(target)

	if normKey != "" && normKey == normTarget {
		return true
	}

	if isDockerHubRegistry(normKey) && isDockerHubRegistry(normTarget) {
		return true
	}

	return false
}

func matchesAnyRegistry(key string, targets []string) bool {
	if len(targets) == 0 {
		return true
	}

	for _, target := range targets {
		if matchesRegistry(key, target) {
			return true
		}
	}

	return false
}

// getDockerAuthConfigs returns a map with the auth configs from the docker config file
// using the registry as the key. If registries are provided, only matching registries are resolved.
func getDockerAuthConfigs(registries ...string) (map[string]registry.AuthConfig, error) {
	cfg, err := getDockerConfig()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]registry.AuthConfig{}, nil
		}

		return nil, err
	}

	key, err := configKey(cfg)
	if err != nil {
		return nil, err
	}

	authConfigsToProcess := make(map[string]dockercfg.AuthConfig)
	for k, v := range cfg.AuthConfigs {
		if matchesAnyRegistry(k, registries) {
			authConfigsToProcess[k] = v
		}
	}

	var credHelpersToProcess []string
	for k := range cfg.CredentialHelpers {
		if matchesAnyRegistry(k, registries) {
			credHelpersToProcess = append(credHelpersToProcess, k)
		}
	}

	size := len(authConfigsToProcess) + len(credHelpersToProcess)
	cfgs := make(map[string]registry.AuthConfig, size)
	if size == 0 {
		return cfgs, nil
	}

	results := make(chan authConfigResult, size)
	var wg sync.WaitGroup
	wg.Add(size)
	for k, v := range authConfigsToProcess {
		go func(k string, v dockercfg.AuthConfig) {
			defer wg.Done()

			ac := registry.AuthConfig{
				Auth:          v.Auth,
				IdentityToken: v.IdentityToken,
				Password:      v.Password,
				RegistryToken: v.RegistryToken,
				ServerAddress: v.ServerAddress,
				Username:      v.Username,
			}

			switch {
			case ac.Username == "" && ac.Password == "":
				// Look up credentials from the credential store.
				if err := creds.AuthConfig(k, key, &ac); err != nil {
					results <- authConfigResult{err: err}
					return
				}
			case ac.Auth == "":
				// Create auth from the username and password encoding.
				ac.Auth = base64.StdEncoding.EncodeToString([]byte(ac.Username + ":" + ac.Password))
			}

			results <- authConfigResult{key: k, cfg: ac}
		}(k, v)
	}

	// In the case where the auth field in the .docker/conf.json is empty, and the user has
	// credential helpers registered the auth comes from there.
	for _, k := range credHelpersToProcess {
		go func(k string) {
			defer wg.Done()

			var ac registry.AuthConfig
			if err := creds.AuthConfig(k, key, &ac); err != nil {
				results <- authConfigResult{err: err}
				return
			}

			results <- authConfigResult{key: k, cfg: ac}
		}(k)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var errs []error
	for result := range results {
		if result.err != nil {
			errs = append(errs, result.err)
			continue
		}

		cfgs[result.key] = result.cfg
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return cfgs, nil
}

// getDockerConfig returns the docker config file. It will internally check, in this particular order:
// 1. the DOCKER_AUTH_CONFIG environment variable, unmarshalling it into a dockercfg.Config
// 2. the DOCKER_CONFIG environment variable, as the path to the config file
// 3. else it will load the default config file, which is ~/.docker/config.json
func getDockerConfig() (*dockercfg.Config, error) {
	if env := os.Getenv("DOCKER_AUTH_CONFIG"); env != "" {
		var cfg dockercfg.Config
		if err := json.Unmarshal([]byte(env), &cfg); err != nil {
			return nil, fmt.Errorf("unmarshal DOCKER_AUTH_CONFIG: %w", err)
		}

		return &cfg, nil
	}

	cfg, err := dockercfg.LoadDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("load default config: %w", err)
	}

	return &cfg, nil
}
