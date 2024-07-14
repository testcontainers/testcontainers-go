package testcontainers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"os"
	"sync"

	"github.com/cpuguy83/dockercfg"
	"github.com/docker/docker/api/types/registry"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// defaultRegistryFn is variable overwritten in tests to check for behaviour with different default values.
var defaultRegistryFn = defaultRegistry

// DockerImageAuth returns the auth config for the given Docker image, extracting first its Docker registry.
// Finally, it will use the credential helpers to extract the information from the docker config file
// for that registry, if it exists.
func DockerImageAuth(ctx context.Context, image string) (string, registry.AuthConfig, error) {
	defaultRegistry := defaultRegistryFn(ctx)
	reg := core.ExtractRegistry(image, defaultRegistry)

	cfgs, err := getDockerAuthConfigs()
	if err != nil {
		return reg, registry.AuthConfig{}, err
	}

	if cfg, ok := getRegistryAuth(reg, cfgs); ok {
		return reg, cfg, nil
	}

	return reg, registry.AuthConfig{}, dockercfg.ErrCredentialsNotFound
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
	client, err := NewDockerClientWithOpts(ctx)
	if err != nil {
		return core.IndexDockerIO
	}
	defer client.Close()

	info, err := client.Info(ctx)
	if err != nil {
		return core.IndexDockerIO
	}

	return info.IndexServerAddress
}

// authConfig represents the details of the auth config for a registry.
type authConfig struct {
	key string
	cfg registry.AuthConfig
}

// getDockerAuthConfigs returns a map with the auth configs from the docker config file
// using the registry as the key
var getDockerAuthConfigs = sync.OnceValues(func() (map[string]registry.AuthConfig, error) {
	cfg, err := getDockerConfig()
	if err != nil {
		return nil, err
	}

	cfgs := map[string]registry.AuthConfig{}
	results := make(chan authConfig, len(cfg.AuthConfigs)+len(cfg.CredentialHelpers))
	var wg sync.WaitGroup
	wg.Add(len(cfg.AuthConfigs) + len(cfg.CredentialHelpers))
	for k, v := range cfg.AuthConfigs {
		go func(k string, v dockercfg.AuthConfig) {
			defer wg.Done()

			ac := registry.AuthConfig{
				Auth:          v.Auth,
				Email:         v.Email,
				IdentityToken: v.IdentityToken,
				Password:      v.Password,
				RegistryToken: v.RegistryToken,
				ServerAddress: v.ServerAddress,
				Username:      v.Username,
			}

			if v.Username == "" && v.Password == "" {
				u, p, _ := dockercfg.GetRegistryCredentials(k)
				ac.Username = u
				ac.Password = p
			}

			if v.Auth == "" {
				ac.Auth = base64.StdEncoding.EncodeToString([]byte(ac.Username + ":" + ac.Password))
			}
			results <- authConfig{key: k, cfg: ac}
		}(k, v)
	}

	// in the case where the auth field in the .docker/conf.json is empty, and the user has credential helpers registered
	// the auth comes from there
	for k := range cfg.CredentialHelpers {
		go func(k string) {
			defer wg.Done()

			ac := registry.AuthConfig{}
			u, p, _ := dockercfg.GetRegistryCredentials(k)
			ac.Username = u
			ac.Password = p
			results <- authConfig{key: k, cfg: ac}
		}(k)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for ac := range results {
		cfgs[ac.key] = ac.cfg
	}

	return cfgs, nil
})

// getDockerConfig returns the docker config file. It will internally check, in this particular order:
// 1. the DOCKER_AUTH_CONFIG environment variable, unmarshalling it into a dockercfg.Config
// 2. the DOCKER_CONFIG environment variable, as the path to the config file
// 3. else it will load the default config file, which is ~/.docker/config.json
func getDockerConfig() (dockercfg.Config, error) {
	dockerAuthConfig := os.Getenv("DOCKER_AUTH_CONFIG")
	if dockerAuthConfig != "" {
		cfg := dockercfg.Config{}
		err := json.Unmarshal([]byte(dockerAuthConfig), &cfg)
		if err == nil {
			return cfg, nil
		}
	}

	cfg, err := dockercfg.LoadDefaultConfig()
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
