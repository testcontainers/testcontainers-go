package testcontainers

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/cpuguy83/dockercfg"
	"github.com/docker/docker/api/types"
)

// AuthFromDockerConfig returns the auth config for the given registry, using the credential helpers
// to extract the information from the docker config file
func AuthFromDockerConfig(registry string) (types.AuthConfig, error) {
	cfgs, err := GetDockerAuthConfigs()
	if err != nil {
		return types.AuthConfig{}, err
	}

	if cfg, ok := cfgs[registry]; ok {
		if cfg.Username == "" && cfg.Password == "" {
			u, p, err := dockercfg.GetRegistryCredentials(registry)
			if err != nil {
				return types.AuthConfig{}, err
			}
			cfg.Username = u
			cfg.Password = p
		}

		cfg.Auth = base64.StdEncoding.EncodeToString([]byte(cfg.Username + ":" + cfg.Password))

		return cfg, nil
	}

	return types.AuthConfig{}, dockercfg.ErrCredentialsNotFound
}

// GetDockerAuthConfigs returns a map with the auth configs from the docker config file
// using the registry as the key
func GetDockerAuthConfigs() (map[string]types.AuthConfig, error) {
	cfg, err := getDockerConfig()
	if err != nil {
		return nil, err
	}

	cfgs := map[string]types.AuthConfig{}
	for k, v := range cfg.AuthConfigs {
		cfgs[k] = types.AuthConfig{
			Auth:          v.Auth,
			Email:         v.Email,
			IdentityToken: v.IdentityToken,
			Password:      v.Password,
			RegistryToken: v.RegistryToken,
			ServerAddress: v.ServerAddress,
			Username:      v.Username,
		}
	}

	return cfgs, nil
}

// getDockerConfig returns the docker config file. It will internally check, in this particular order:
// 1. the DOCKER_AUTH_CONFIG environment variable, unmarshalling it into a dockercfg.Config
// 2. the DOCKER_CONFIG environment variable, as the path to the config file
// 3. else it will load the default config file, which is ~/.docker/config.json
func getDockerConfig() (dockercfg.Config, error) {
	dockerAuthConfig := os.Getenv("DOCKER_AUTH_CONFIG")
	if dockerAuthConfig != "" {
		cfg := dockercfg.Config{}
		err := loadFromReader([]byte(dockerAuthConfig), &cfg)
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

// loadFromReader loads config from the specified path into cfg
func loadFromReader(config []byte, cfg *dockercfg.Config) error {
	return json.Unmarshal(config, &cfg)
}
