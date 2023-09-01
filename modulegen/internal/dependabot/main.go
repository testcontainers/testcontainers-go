package dependabot

import (
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

// update examples in dependabot
func GenerateDependabotUpdates(ctx *context.Context, m context.TestcontainersModule) error {
	directory := "/" + m.ParentDir() + "/" + m.Lower()
	return UpdateConfig(ctx.DependabotConfigFile(), directory, "gomod")
}

func UpdateConfig(configFile string, directory string, packageEcosystem string) error {
	config, err := readConfig(configFile)
	if err != nil {
		return err
	}
	config.addUpdate(newUpdate(directory, packageEcosystem))
	return writeConfig(configFile, config)
}

func GetUpdates(configFile string) (Updates, error) {
	config, err := readConfig(configFile)
	if err != nil {
		return nil, err
	}
	return config.Updates, nil
}

func CopyConfig(configFile string, tmpFile string) error {
	config, err := readConfig(configFile)
	if err != nil {
		return err
	}
	return writeConfig(tmpFile, config)
}
