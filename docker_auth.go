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

// getRegistryCredentials is a variable overwritten in tests to mock the credentials lookup.
// It resolves against the config already loaded, which honours DOCKER_AUTH_CONFIG, rather than
// reloading the default config file.
var getRegistryCredentials = func(cfg *dockercfg.Config, hostname string) (string, string, error) {
	return cfg.GetRegistryCredentials(hostname)
}

// DockerImageAuth returns the auth config for the given Docker image, extracting first its Docker registry.
// Finally, it will use the credential helpers to extract the information from the docker config file
// for that registry, if it exists.
func DockerImageAuth(ctx context.Context, image string) (string, registry.AuthConfig, error) {
	cfg, configs, err := loadDockerAuth()
	if err != nil {
		reg := core.ExtractRegistry(image, defaultRegistryFn(ctx))
		return reg, registry.AuthConfig{}, err
	}

	return dockerImageAuth(ctx, image, cfg, configs)
}

// loadDockerAuth loads the docker config once and the auth configs derived from it,
// so that both the map lookup and the credentials store lookup see the same config.
// A missing config file is not an error: it yields an empty config.
func loadDockerAuth() (*dockercfg.Config, map[string]registry.AuthConfig, error) {
	cfg, err := getDockerConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, nil, err
		}
		cfg = &dockercfg.Config{}
	}

	configs, err := getDockerAuthConfigsFromConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	return cfg, configs, nil
}

// dockerImageAuth returns the auth config for the given Docker image.
func dockerImageAuth(ctx context.Context, image string, cfg *dockercfg.Config, configs map[string]registry.AuthConfig) (string, registry.AuthConfig, error) {
	defaultRegistry := defaultRegistryFn(ctx)
	reg := core.ExtractRegistry(image, defaultRegistry)

	// Normalize Docker Hub aliases for credential lookup
	if strings.EqualFold(reg, "docker.io") ||
		strings.EqualFold(reg, "registry.hub.docker.com") ||
		strings.EqualFold(reg, "registry-1.docker.io") {
		reg = defaultRegistry // This is https://index.docker.io/v1/
	}

	if ac, ok := getRegistryAuth(reg, configs); ok {
		return reg, ac, nil
	}

	ac, ok, err := credentialsStoreAuth(cfg, reg)
	if err != nil {
		return reg, registry.AuthConfig{}, err
	}
	if ok {
		return reg, ac, nil
	}

	return reg, registry.AuthConfig{}, dockercfg.ErrCredentialsNotFound
}

// credentialsStoreAuth returns the auth config the credentials store holds for reg,
// if one is configured and it has an entry for that registry.
//
// A credentials store serves every registry, so unlike auths and credHelpers it has
// no entries to enumerate up front: it can only be asked once the registry is known.
// See https://docs.docker.com/reference/cli/docker/login/#credential-stores
func credentialsStoreAuth(cfg *dockercfg.Config, reg string) (registry.AuthConfig, bool, error) {
	// A store cannot be asked about an empty host, and a missing helper binary
	// already reads as "no credentials" further down.
	if cfg.CredentialsStore == "" || reg == "" {
		return registry.AuthConfig{}, false, nil
	}

	key, err := configKey(cfg)
	if err != nil {
		return registry.AuthConfig{}, false, err
	}

	var ac registry.AuthConfig
	if err := creds.AuthConfig(cfg, reg, key, &ac); err != nil {
		return registry.AuthConfig{}, false, err
	}

	// The store reports an unknown registry as empty credentials rather than an error.
	if ac.Username == "" && ac.Password == "" && ac.IdentityToken == "" {
		return registry.AuthConfig{}, false, nil
	}

	return ac, true, nil
}

func getRegistryAuth(reg string, cfgs map[string]registry.AuthConfig) (registry.AuthConfig, bool) {
	if cfg, ok := cfgs[reg]; ok {
		return cfg, true
	}

	// fallback match using authentication key host
	for k, cfg := range cfgs {
		keyURL, err := url.Parse(k)
		if err != nil {
			continue
		}

		host := keyURL.Host
		if keyURL.Scheme == "" {
			// url.Parse: The url may be relative (a path, without a host) [...]
			host = keyURL.Path
		}

		if host == reg {
			return cfg, true
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
func (c *credentialsCache) AuthConfig(cfg *dockercfg.Config, hostname, configKey string, authConfig *registry.AuthConfig) error {
	u, p, err := creds.get(cfg, hostname, configKey)
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
func (c *credentialsCache) get(cfg *dockercfg.Config, hostname, configKey string) (string, string, error) {
	key := configKey + ":" + hostname
	c.mtx.RLock()
	entry, ok := c.entries[key]
	c.mtx.RUnlock()

	if ok {
		return entry.username, entry.password, nil
	}

	// No entry found, request and cache.
	user, password, err := getRegistryCredentials(cfg, hostname)
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

// getDockerAuthConfigs returns a map with the auth configs from the docker config file
// using the registry as the key
func getDockerAuthConfigs() (map[string]registry.AuthConfig, error) {
	cfg, err := getDockerConfig()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]registry.AuthConfig{}, nil
		}

		return nil, err
	}

	return getDockerAuthConfigsFromConfig(cfg)
}

// getDockerAuthConfigsFromConfig returns a map with the auth configs from the given docker config
// using the registry as the key
func getDockerAuthConfigsFromConfig(cfg *dockercfg.Config) (map[string]registry.AuthConfig, error) {
	key, err := configKey(cfg)
	if err != nil {
		return nil, err
	}

	size := len(cfg.AuthConfigs) + len(cfg.CredentialHelpers)
	cfgs := make(map[string]registry.AuthConfig, size)
	results := make(chan authConfigResult, size)
	var wg sync.WaitGroup
	wg.Add(size)
	for k, v := range cfg.AuthConfigs {
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
				if err := creds.AuthConfig(cfg, k, key, &ac); err != nil {
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
	for k := range cfg.CredentialHelpers {
		go func(k string) {
			defer wg.Done()

			var ac registry.AuthConfig
			if err := creds.AuthConfig(cfg, k, key, &ac); err != nil {
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
