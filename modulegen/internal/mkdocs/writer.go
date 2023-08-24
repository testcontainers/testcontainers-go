package mkdocs

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func writeConfig(configFile string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// simple solution to replace the empty strings, as mapping those fields
	// into the MkDocs config is not supported yet
	content := string(data)
	content = strings.ReplaceAll(content, `emoji_generator: ""`, "emoji_generator: !!python/name:materialx.emoji.to_svg")
	content = strings.ReplaceAll(content, `emoji_index: ""`, "emoji_index: !!python/name:materialx.emoji.twemoji")

	data = []byte(content)

	return os.WriteFile(configFile, data, 0o777)
}
