package testcontainers

import (
	"github.com/cpuguy83/dockercfg"
	"github.com/docker/docker/api/types"
)

// AuthFromDockerConfig returns the Base64 auth string for a registry from the docker config file
func AuthFromDockerConfig(registry string) (types.AuthConfig, error) {
	cfgs, err := GetDockerAuthConfigs()
	if err != nil {
		return types.AuthConfig{}, err
	}

	if cfg, ok := cfgs[registry]; ok {
		return cfg, nil
	}

	return types.AuthConfig{}, dockercfg.ErrCredentialsNotFound
}

// GetDockerAuthConfigs returns a map with the auth configs from the docker config file
// using the registry as the key
func GetDockerAuthConfigs() (map[string]types.AuthConfig, error) {
	cfg, err := dockercfg.LoadDefaultConfig()
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
