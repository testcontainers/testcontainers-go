package mkdocs

import (
	"os"

	"gopkg.in/yaml.v3"
)

func writeConfig(configFile string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0o777)
}
