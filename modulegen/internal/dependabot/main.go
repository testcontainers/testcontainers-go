package dependabot

import "github.com/testcontainers/testcontainers-go/modulegen/internal/context"

type Generator struct{}

// AddModule update dependabot with the new module
func (g Generator) AddModule(ctx context.Context, tcModule context.TestcontainersModule) error {
	configFile := ctx.DependabotConfigFile()

	config, err := readConfig(configFile)
	if err != nil {
		return err
	}

	packageEcosystem := "gomod"
	directory := "/" + tcModule.ParentDir() + "/" + tcModule.Lower()

	config.addUpdate(newUpdate(directory, packageEcosystem))

	return writeConfig(configFile, config)
}

// Generate generates dependabot config file from source
func (g Generator) Generate(ctx context.Context) error {
	configFile := ctx.DependabotConfigFile()

	config, err := readConfig(configFile)
	if err != nil {
		return err
	}

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
