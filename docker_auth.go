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
		ac := types.AuthConfig{
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
			v.Username = u
			v.Password = p
		}

		v.Auth = base64.StdEncoding.EncodeToString([]byte(v.Username + ":" + v.Password))

		cfgs[k] = ac
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
