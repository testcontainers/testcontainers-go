package dependabot

import (
	"os"

	"gopkg.in/yaml.v3"
)

func readConfig(configFile string) (*Config, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
