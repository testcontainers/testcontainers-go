package testcontainers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/cpuguy83/dockercfg"
	"github.com/docker/docker/api/types"
)

const indexDockerIO = "https://index.docker.io/v1/"

// RegistryAuth returns the auth config for the given registry, using the credential helpers
// to extract the information from the docker config file
func RegistryAuth(registry string) (types.AuthConfig, error) {
	cfgs, err := getDockerAuthConfigs()
	if err != nil {
		return types.AuthConfig{}, err
	}

	if cfg, ok := cfgs[registry]; ok {
		return cfg, nil
	}

	return types.AuthConfig{}, dockercfg.ErrCredentialsNotFound
}

// DefaultRegistryAuth returns the auth config for the default registry, using the credential helpers
// to extract the information from the docker config file
func DefaultRegistryAuth(ctx context.Context) (types.AuthConfig, error) {
	return RegistryAuth(defaultRegistry(ctx))
}

// defaultRegistry returns the default registry to use when pulling images
// It will use the docker daemon to get the default registry, returning "https://index.docker.io/v1/" if
// it fails to get the information from the daemon
func defaultRegistry(ctx context.Context) string {
	p, err := NewDockerProvider()
	if err != nil {
		return indexDockerIO
	}

	info, err := p.client.Info(ctx)
	if err != nil {
		return indexDockerIO
	}

	return info.IndexServerAddress
}

// getDockerAuthConfigs returns a map with the auth configs from the docker config file
// using the registry as the key
func getDockerAuthConfigs() (map[string]types.AuthConfig, error) {
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
