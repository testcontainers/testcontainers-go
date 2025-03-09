package dependabot

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func readConfig(configFile string) (*Config, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	config := &Config{}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal file: %w", err)
	}

	return config, nil
}
