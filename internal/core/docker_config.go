package core

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cpuguy83/dockercfg"
)

// ReadDockerConfig returns the docker config file. It will internally check, in this particular order:
// 1. the DOCKER_AUTH_CONFIG environment variable, unmarshalling it into a dockercfg.Config
// 2. the DOCKER_CONFIG environment variable, as the path to the config file
// 3. else it will load the default config file, which is ~/.docker/config.json
func ReadDockerConfig() (*dockercfg.Config, error) {
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
